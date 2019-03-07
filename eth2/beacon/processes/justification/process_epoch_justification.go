package justification

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

func ProcessEpochCrosslinks(state *beacon.BeaconState) {

	current_epoch := state.Epoch()
	// epoch numbers are trusted, no errors
	previous_boundary_block_root, _ := state.Get_block_root((current_epoch - 1).GetStartSlot())
	current_boundary_block_root, _ := state.Get_block_root(current_epoch.GetStartSlot())

	previous_epoch_boundary_attester_indices := make([]beacon.ValidatorIndex, 0)
	current_epoch_boundary_attester_indices := make([]beacon.ValidatorIndex, 0)
	for _, att := range state.PreviousEpochAttestations {
		// If the attestation is for the boundary:
		if att.Data.Epoch_boundary_root == previous_boundary_block_root {
			participants, _ := state.Get_attestation_participants(&att.Data, &att.Aggregation_bitfield)
			for _, vIndex := range participants {
				previous_epoch_boundary_attester_indices = append(previous_epoch_boundary_attester_indices, vIndex)
			}
		}
	}
	for _, att := range state.CurrentEpochAttestations {
		// If the attestation is for the boundary:
		if att.Data.Epoch_boundary_root == current_boundary_block_root {
			participants, _ := state.Get_attestation_participants(&att.Data, &att.Aggregation_bitfield)
			for _, vIndex := range participants {
				current_epoch_boundary_attester_indices = append(current_epoch_boundary_attester_indices, vIndex)
			}
		}
	}

	new_justified_epoch := state.Justified_epoch
	// Rotate the justification bitfield up one epoch to make room for the current epoch
	state.Justification_bitfield <<= 1

	// Get the sum balances of the boundary attesters, and the total balance at the time.
	previous_epoch_boundary_attesting_balance := state.Validator_balances.Get_total_balance(previous_epoch_boundary_attester_indices)
	previous_total_balance := state.Validator_balances.Get_total_balance(state.Validator_registry.Get_active_validator_indices(current_epoch - 1))
	current_epoch_boundary_attesting_balance := state.Validator_balances.Get_total_balance(current_epoch_boundary_attester_indices)
	current_total_balance := state.Validator_balances.Get_total_balance(state.Validator_registry.Get_active_validator_indices(current_epoch))

	// > Justification
	// If the previous epoch gets justified, fill the second last bit
	if 3*previous_epoch_boundary_attesting_balance >= 2*previous_total_balance {
		state.Justification_bitfield |= 2
		new_justified_epoch = current_epoch - 1
	}
	// If the current epoch gets justified, fill the last bit
	if 3*current_epoch_boundary_attesting_balance >= 2*current_total_balance {
		state.Justification_bitfield |= 1
		new_justified_epoch = current_epoch
	}
	// > Finalization
	// The 2nd/3rd/4th most recent epochs are all justified, the 2nd using the 4th as source
	if (state.Justification_bitfield>>1)&7 == 7 && state.Previous_justified_epoch == current_epoch-3 {
		state.Finalized_epoch = state.Previous_justified_epoch
	}
	// The 2nd/3rd most recent epochs are both justified, the 2nd using the 3rd as source
	if (state.Justification_bitfield>>1)&3 == 3 && state.Previous_justified_epoch == current_epoch-2 {
		state.Finalized_epoch = state.Previous_justified_epoch
	}
	// The 1st/2nd/3rd most recent epochs are all justified, the 1st using the 3rd as source
	if (state.Justification_bitfield>>0)&7 == 7 && state.Justified_epoch == current_epoch-2 {
		state.Finalized_epoch = state.Justified_epoch
	}
	// The 1st/2nd most recent epochs are both justified, the 1st using the 2nd as source
	if (state.Justification_bitfield>>0)&3 == 3 && state.Justified_epoch == current_epoch-1 {
		state.Finalized_epoch = state.Justified_epoch
	}
	// Rotate justified epochs
	state.Previous_justified_epoch = state.Justified_epoch
	state.Justified_epoch = new_justified_epoch
}
