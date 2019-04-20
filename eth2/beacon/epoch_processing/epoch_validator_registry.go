package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"sort"
)

func ProcessEpochValidatorRegistry(state *BeaconState) {
	activationQueue := make([]ValidatorIndex, 0)
	for i, v := range state.ValidatorRegistry {
		if v.ActivationEligibilityEpoch != FAR_FUTURE_EPOCH &&
			v.ActivationEpoch >= state.FinalizedEpoch.GetDelayedActivationExitEpoch() {
			activationQueue = append(activationQueue, ValidatorIndex(i))
		}
	}
	sort.Slice(activationQueue, func(i int, j int) bool {
		return state.ValidatorRegistry[activationQueue[i]].ActivationEligibilityEpoch <
			state.ValidatorRegistry[activationQueue[j]].ActivationEligibilityEpoch
	})
	for i := uint64(0); i < state.GetChurnLimit(); i++ {
		state.ActivateValidator(activationQueue[i], false)
	}
	state.LatestStartShard = (
		state.LatestStartShard +
		Shard(state.GetShardDelta(state.Epoch()))) % SHARD_COUNT
}
