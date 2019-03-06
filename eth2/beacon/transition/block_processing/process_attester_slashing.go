package block_processing

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/transition"
)

func ProcessAttesterSlashings(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Attester_slashings) > eth2.MAX_ATTESTER_SLASHINGS {
		return errors.New("too many attester slashings")
	}
	for _, attester_slashing := range block.Body.Attester_slashings {
		if err := ProcessAttesterSlashing(state, &attester_slashing); err != nil {
			return err
		}
	}
}

func ProcessAttesterSlashing(state *beacon.BeaconState, attester_slashing *beacon.AttesterSlashing) error {
	sa1, sa2 := &attester_slashing.Slashable_attestation_1, &attester_slashing.Slashable_attestation_2
	// verify the attester_slashing
	if !(sa1.Data != sa2.Data && (
			transition.Is_double_vote(&sa1.Data, &sa2.Data) ||
			transition.Is_surround_vote(&sa1.Data, &sa2.Data)) &&
		transition.Verify_slashable_attestation(state, sa1) &&
		transition.Verify_slashable_attestation(state, sa2)) {
		return errors.New("attester slashing is invalid")
	}
	// keep track of effectiveness
	slashedAny := false
	// run slashings where applicable

	// indices are trusted, they have been verified by verify_slashable_attestation(...)
	for _, v1 := range sa1.Validator_indices {
		for _, v2 := range sa2.Validator_indices {
			if v1 == v2 && !state.Validator_registry[v1].Slashed {
				if err := transition.Slash_validator(state, v1); err != nil {
					return err
				}
				slashedAny = true
				continue
			}
		}
	}
	// "Verify that len(slashable_indices) >= 1."
	if !slashedAny {
		return errors.New("attester slashing %d is not effective, hence invalid")
	}
}
