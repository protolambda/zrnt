package slashings

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/components/attestations"
	. "github.com/protolambda/zrnt/eth2/beacon/components/meta"
	. "github.com/protolambda/zrnt/eth2/core"
)

type AttesterSlashingReq interface {
	RegistrySizeMeta
	PubkeyMeta
	VersioningMeta
	ValidatorMeta
	ProposingMeta
	BalanceMeta
	ExitMeta
}

func (state *SlashingsState) ProcessAttesterSlashings(meta AttesterSlashingReq, ops []AttesterSlashing) error {
	for i := range ops {
		if err := state.ProcessAttesterSlashing(meta, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

type AttesterSlashing struct {
	Attestation1 IndexedAttestation
	Attestation2 IndexedAttestation
}

func (state *SlashingsState) ProcessAttesterSlashing(meta AttesterSlashingReq, attesterSlashing *AttesterSlashing) error {
	sa1 := &attesterSlashing.Attestation1
	sa2 := &attesterSlashing.Attestation2

	if !IsSlashableAttestationData(&sa1.Data, &sa2.Data) {
		return errors.New("attester slashing is has no valid reasoning")
	}

	if err := sa1.Validate(meta); err != nil {
		return errors.New("attestation 1 of attester slashing cannot be verified")
	}
	if err := sa2.Validate(meta); err != nil {
		return errors.New("attestation 2 of attester slashing cannot be verified")
	}

	// keep track of effectiveness
	slashedAny := false

	// the individual custody index sets are already sorted (as verified by ValidateIndexedAttestation)
	// just merge them in O(n)
	indices1 := ValidatorSet(sa1.CustodyBit0Indices).MergeDisjoint(ValidatorSet(sa1.CustodyBit1Indices))
	indices2 := ValidatorSet(sa2.CustodyBit0Indices).MergeDisjoint(ValidatorSet(sa2.CustodyBit1Indices))

	currentEpoch := meta.Epoch()
	// run slashings where applicable
	indices1.ZigZagJoin(indices2, func(i ValidatorIndex) {
		if meta.Validator(i).IsSlashable(currentEpoch) {
			state.SlashValidator(meta, i, nil)
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
