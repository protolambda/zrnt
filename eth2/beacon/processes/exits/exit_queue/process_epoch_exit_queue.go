package exit_queue

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"sort"
)

func ProcessEpochExitQueue(state *beacon.BeaconState) {
	current_epoch := state.Epoch()
	eligible_indices := make([]beacon.ValidatorIndex, 0)
	for index, validator := range state.Validator_registry {
		// Filter out dequeued validators
		if validator.Withdrawable_epoch == beacon.FAR_FUTURE_EPOCH {
			continue
		}
		// Dequeue if the minimum amount of time has passed
		if current_epoch > validator.Exit_epoch+beacon.MIN_VALIDATOR_WITHDRAWABILITY_DELAY {
			eligible_indices = append(eligible_indices, beacon.ValidatorIndex(index))
		}
	}
	// Sort in order of exit epoch, and validators that exit within the same epoch exit in order of validator index
	sort.Slice(eligible_indices, func(i int, j int) bool {
		return state.Validator_registry[eligible_indices[i]].Exit_epoch < state.Validator_registry[eligible_indices[j]].Exit_epoch
	})
	for i, end := uint64(0), uint64(len(eligible_indices)); i < beacon.MAX_EXIT_DEQUEUES_PER_EPOCH && i < end; i++ {
		state.Prepare_validator_for_withdrawal(eligible_indices[i])
	}
}
