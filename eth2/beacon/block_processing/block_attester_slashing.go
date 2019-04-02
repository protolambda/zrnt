package block_processing

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/beacon"
)

func ProcessBlockAttesterSlashings(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.AttesterSlashings) > beacon.MAX_ATTESTER_SLASHINGS {
		return errors.New("too many attester slashings")
	}
	for _, attesterSlashing := range block.Body.AttesterSlashings {
		if err := ProcessAttesterSlashing(state, &attesterSlashing); err != nil {
			return err
		}
	}
	return nil
}

func ProcessAttesterSlashing(state *beacon.BeaconState, attesterSlashing *beacon.AttesterSlashing) error {
	sa1 := &attesterSlashing.Attestation1
	sa2 := &attesterSlashing.Attestation2

	// check that the attestations are conflicting
	if sa1.Data != sa2.Data {
		return errors.New("attestations of attester slashing are not conflicting")
	}

	// verify the attester_slashing
	if !(IsDoubleVote(&sa1.Data, &sa2.Data) || IsSurroundVote(&sa1.Data, &sa2.Data)) {
		return errors.New("attester slashing is has no valid reasoning")
	}
	if !state.VerifyIndexedAttestation(sa1) {
		return errors.New("attestation 1 of attester slashing cannot be verified")
	}
	if !state.VerifyIndexedAttestation(sa2) {
		return errors.New("attestation 2 of attester slashing cannot be verified")
	}

	// keep track of effectiveness
	slashedAny := false

	// indices are trusted (valid range), they have been verified by verify_slashable_attestation(...)
	indices1 := make([]beacon.ValidatorIndex, 0, len(sa1.CustodyBit0Indexes) + len(sa1.CustodyBit1Indexes))
	indices1 = append(indices1, sa1.CustodyBit0Indexes...)
	indices1 = append(indices1, sa1.CustodyBit1Indexes...)
	indices2 := make([]beacon.ValidatorIndex, 0, len(sa2.CustodyBit0Indexes) + len(sa2.CustodyBit1Indexes))
	indices2 = append(indices1, sa2.CustodyBit0Indexes...)
	indices2 = append(indices1, sa2.CustodyBit1Indexes...)

	// run slashings where applicable
	intersection, _ := beacon.FindInAndOutValidators(indices1, indices2)
	for _, v := range intersection {
		if !state.ValidatorRegistry[v].Slashed {
			if err := state.SlashValidator(v); err != nil {
				return err
			}
			slashedAny = true
			continue
		}
	}
	// "Verify that len(slashable_indices) >= 1."
	if !slashedAny {
		return errors.New("attester slashing %d is not effective, hence invalid")
	}
	return nil
}

// Check if a and b have the same target epoch.
func IsDoubleVote(a *beacon.AttestationData, b *beacon.AttestationData) bool {
	return a.Slot.ToEpoch() == b.Slot.ToEpoch()
}

// Check if a surrounds b, i.E. source(a) < source(b) and target(a) > target(b)
func IsSurroundVote(a *beacon.AttestationData, b *beacon.AttestationData) bool {
	return a.SourceEpoch < b.SourceEpoch && a.Slot.ToEpoch() > b.Slot.ToEpoch()
}
