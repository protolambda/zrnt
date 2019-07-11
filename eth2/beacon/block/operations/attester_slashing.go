package operations

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
)

type AttesterSlashings []AttesterSlashing

func (_ *AttesterSlashings) Limit() uint32 {
	return MAX_ATTESTER_SLASHINGS
}

func (ops AttesterSlashings) Process(state *BeaconState) error {
	for _, op := range ops {
		if err := op.Process(state); err != nil {
			return err
		}
	}
	return nil
}

type AttesterSlashing struct {
	// First attestation
	Attestation1 IndexedAttestation
	// Second attestation
	Attestation2 IndexedAttestation
}

func (attesterSlashing *AttesterSlashing) Process(state *BeaconState) error {
	sa1 := &attesterSlashing.Attestation1
	sa2 := &attesterSlashing.Attestation2

	if !IsSlashableAttestationData(&sa1.Data, &sa2.Data) {
		return errors.New("attester slashing is has no valid reasoning")
	}

	if err := state.ValidateIndexedAttestation(sa1); err != nil {
		return errors.New("attestation 1 of attester slashing cannot be verified")
	}
	if err := state.ValidateIndexedAttestation(sa2); err != nil {
		return errors.New("attestation 2 of attester slashing cannot be verified")
	}

	// keep track of effectiveness
	slashedAny := false

	// the individual custody index sets are already sorted (as verified by ValidateIndexedAttestation)
	// just merge them in O(n)
	indices1 := ValidatorSet(sa1.CustodyBit0Indices).MergeDisjoint(ValidatorSet(sa1.CustodyBit1Indices))
	indices2 := ValidatorSet(sa2.CustodyBit0Indices).MergeDisjoint(ValidatorSet(sa2.CustodyBit1Indices))

	currentEpoch := state.Epoch()
	// run slashings where applicable
	indices1.ZigZagJoin(indices2, func(i ValidatorIndex) {
		if state.Validators[i].IsSlashable(currentEpoch) {
			state.SlashValidator(i, nil)
			slashedAny = true
		}
	}, nil)
	if !slashedAny {
		return errors.New("attester slashing %d is not effective, hence invalid")
	}
	return nil
}

func IsSlashableAttestationData(a *AttestationData, b *AttestationData) bool {
	return IsSurroundVote(a, b) || IsDoubleVote(a, b)
}

// Check if a and b have the same target epoch.
func IsDoubleVote(a *AttestationData, b *AttestationData) bool {
	return *a != *b && a.Target.Epoch == b.Target.Epoch
}

// Check if a surrounds b, i.E. source(a) < source(b) and target(a) > target(b)
func IsSurroundVote(a *AttestationData, b *AttestationData) bool {
	return a.Source.Epoch < b.Source.Epoch && a.Target.Epoch > b.Target.Epoch
}
