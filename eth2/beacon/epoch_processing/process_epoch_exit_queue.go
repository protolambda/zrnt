package epoch_processing

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"sort"
)

func ProcessEpochExitQueue(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	eligibleIndices := make([]beacon.ValidatorIndex, 0)
	for index, validator := range state.ValidatorRegistry {
		// Filter out dequeued validators
		if validator.WithdrawableEpoch == beacon.FAR_FUTURE_EPOCH {
			continue
		}
		// Dequeue if the minimum amount of time has passed
		if currentEpoch > validator.ExitEpoch+beacon.MIN_VALIDATOR_WITHDRAWABILITY_DELAY {
			eligibleIndices = append(eligibleIndices, beacon.ValidatorIndex(index))
		}
	}
	// Sort in order of exit epoch, and validators that exit within the same epoch exit in order of validator index
	sort.Slice(eligibleIndices, func(i int, j int) bool {
		return state.ValidatorRegistry[eligibleIndices[i]].ExitEpoch < state.ValidatorRegistry[eligibleIndices[j]].ExitEpoch
	})
	for i, end := uint64(0), uint64(len(eligibleIndices)); i < beacon.MAX_EXIT_DEQUEUES_PER_EPOCH && i < end; i++ {
		state.PrepareValidatorForWithdrawal(eligibleIndices[i])
	}
}
