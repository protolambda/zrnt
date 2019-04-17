package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"sort"
)

func ProcessEpochValidatorRegistry(state *beacon.BeaconState) {
	activationQueue := make([]beacon.ValidatorIndex, 0)
	for i, v := range state.ValidatorRegistry {
		if v.ActivationEligibilityEpoch != beacon.FAR_FUTURE_EPOCH &&
			v.ActivationEpoch >= state.FinalizedEpoch.GetDelayedActivationExitEpoch() {
			activationQueue = append(activationQueue, beacon.ValidatorIndex(i))
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
		beacon.Shard(state.GetShardDelta(state.Epoch()))) % beacon.SHARD_COUNT
}
