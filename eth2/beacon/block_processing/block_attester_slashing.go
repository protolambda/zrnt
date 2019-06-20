package block_processing

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"sort"
)

func ProcessBlockAttesterSlashings(state *BeaconState, block *BeaconBlock) error {
	if len(block.Body.AttesterSlashings) > MAX_ATTESTER_SLASHINGS {
		return errors.New("too many attester slashings")
	}
	for _, attesterSlashing := range block.Body.AttesterSlashings {
		if err := ProcessAttesterSlashing(state, &attesterSlashing); err != nil {
			return err
		}
	}
	return nil
}

func ProcessAttesterSlashing(state *BeaconState, attesterSlashing *AttesterSlashing) error {
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

	// indices are trusted (valid range), they have been verified by verify_slashable_attestation(...)
	indices1 := make(ValidatorSet, 0, len(sa1.CustodyBit0Indices)+len(sa1.CustodyBit1Indices))
	indices1 = append(indices1, sa1.CustodyBit0Indices...)
	indices1 = append(indices1, sa1.CustodyBit1Indices...)
	sort.Sort(indices1)
	indices2 := make(ValidatorSet, 0, len(sa2.CustodyBit0Indices)+len(sa2.CustodyBit1Indices))
	indices2 = append(indices2, sa2.CustodyBit0Indices...)
	indices2 = append(indices2, sa2.CustodyBit1Indices...)
	sort.Sort(indices2)

	currentEpoch := state.Epoch()
	// run slashings where applicable
	indices1.ZigZagJoin(indices2, func(i ValidatorIndex) {
		if state.ValidatorRegistry[i].IsSlashable(currentEpoch) {
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
	return *a != *b && a.TargetEpoch == b.TargetEpoch
}

// Check if a surrounds b, i.E. source(a) < source(b) and target(a) > target(b)
func IsSurroundVote(a *AttestationData, b *AttestationData) bool {
	return a.SourceEpoch < b.SourceEpoch && a.TargetEpoch > b.TargetEpoch
}
