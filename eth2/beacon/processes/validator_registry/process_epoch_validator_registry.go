package validator_registry

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/math"
)

func ProcessEpochValidatorRegistry(state *beacon.BeaconState) {
	// TODO
}

// TODO: updated quickly, review
func UpdateRegistry(state *beacon.BeaconState) {
	current_epoch := state.Epoch()
	next_epoch := current_epoch + 1
	state.Previous_shuffling_epoch = state.Current_shuffling_epoch
	state.Previous_shuffling_start_shard = state.Current_shuffling_start_shard
	state.Previous_shuffling_seed = state.Current_shuffling_seed

	if state.Finalized_epoch > state.Validator_registry_update_epoch {
		needsUpdate := true
		{
			committee_count := beacon.Get_epoch_committee_count(state.Validator_registry.Get_active_validator_count(current_epoch))
			for i := uint64(0); i < committee_count; i++ {
				if shard := (state.Current_shuffling_start_shard + beacon.Shard(i)) % beacon.SHARD_COUNT; state.Latest_crosslinks[shard].Epoch <= state.Validator_registry_update_epoch {
					needsUpdate = false
				}
			}
		}
		if needsUpdate {
			Update_validator_registry(state)
			state.Current_shuffling_epoch = next_epoch
			// recompute committee count, some validators may not be active anymore due to the above update.
			committee_count := beacon.Get_epoch_committee_count(state.Validator_registry.Get_active_validator_count(current_epoch))
			state.Current_shuffling_start_shard = (state.Current_shuffling_start_shard + beacon.Shard(committee_count)) % beacon.SHARD_COUNT
			// ignore error, current_shuffling_epoch is a trusted input
			state.Current_shuffling_seed = state.Generate_seed(state.Current_shuffling_epoch)
		} else {
			// If a validator registry update does not happen:
			epochs_since_last_registry_update := current_epoch - state.Validator_registry_update_epoch
			if epochs_since_last_registry_update > 1 && math.Is_power_of_two(uint64(epochs_since_last_registry_update)) {
				state.Current_shuffling_epoch = next_epoch
				// Note that state.Current_shuffling_start_shard is left unchanged
				state.Current_shuffling_seed = state.Generate_seed(state.Current_shuffling_epoch)
			}
		}
	}
}


func Update_validator_registry(state *beacon.BeaconState) {
	// TODO
}