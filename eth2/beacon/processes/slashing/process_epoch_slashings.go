package slashing

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)


func ProcessEpochSlashings(state *beacon.BeaconState) {
	current_epoch := state.Epoch()
	active_validator_indices := state.Validator_registry.Get_active_validator_indices(current_epoch)
	total_balance := state.Validator_balances.Get_total_balance(active_validator_indices)

	for index, validator := range state.Validator_registry {
		if validator.Slashed &&
			current_epoch == validator.Withdrawable_epoch-(beacon.LATEST_SLASHED_EXIT_LENGTH/2) {
			epoch_index := current_epoch % beacon.LATEST_SLASHED_EXIT_LENGTH
			total_at_start := state.Latest_slashed_balances[(epoch_index+1)%beacon.LATEST_SLASHED_EXIT_LENGTH]
			total_at_end := state.Latest_slashed_balances[epoch_index]
			balance := state.Validator_balances.Get_effective_balance(beacon.ValidatorIndex(index))
			state.Validator_balances[index] -= beacon.Max(
				balance*beacon.Min(
				(total_at_end-total_at_start)*3,
					total_balance,
				)/total_balance,
				balance/beacon.MIN_PENALTY_QUOTIENT)
		}
	}
}

