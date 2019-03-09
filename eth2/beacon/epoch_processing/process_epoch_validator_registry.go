package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/math"
)

func ProcessEpochValidatorRegistry(state *beacon.BeaconState) {
	// TODO
}

// TODO: updated quickly, review
func UpdateRegistry(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	nextEpoch := currentEpoch + 1
	state.PreviousShufflingEpoch = state.CurrentShufflingEpoch
	state.PreviousShufflingStartShard = state.CurrentShufflingStartShard
	state.PreviousShufflingSeed = state.CurrentShufflingSeed

	if state.FinalizedEpoch > state.ValidatorRegistryUpdateEpoch {
		needsUpdate := true
		{
			committeeCount := beacon.GetEpochCommitteeCount(state.ValidatorRegistry.GetActiveValidatorCount(currentEpoch))
			for i := uint64(0); i < committeeCount; i++ {
				if shard := (state.CurrentShufflingStartShard + beacon.Shard(i)) % beacon.SHARD_COUNT; state.LatestCrosslinks[shard].Epoch <= state.ValidatorRegistryUpdateEpoch {
					needsUpdate = false
				}
			}
		}
		if needsUpdate {
			UpdateValidatorRegistry(state)
			state.CurrentShufflingEpoch = nextEpoch
			// recompute committee count, some validators may not be active anymore due to the above update.
			committeeCount := beacon.GetEpochCommitteeCount(state.ValidatorRegistry.GetActiveValidatorCount(currentEpoch))
			state.CurrentShufflingStartShard = (state.CurrentShufflingStartShard + beacon.Shard(committeeCount)) % beacon.SHARD_COUNT
			// ignore error, current_shuffling_epoch is a trusted input
			state.CurrentShufflingSeed = state.GenerateSeed(state.CurrentShufflingEpoch)
		} else {
			// If a validator registry update does not happen:
			epochsSinceLastRegistryUpdate := currentEpoch - state.ValidatorRegistryUpdateEpoch
			if epochsSinceLastRegistryUpdate > 1 && math.IsPowerOfTwo(uint64(epochsSinceLastRegistryUpdate)) {
				state.CurrentShufflingEpoch = nextEpoch
				// Note that state.Current_shuffling_start_shard is left unchanged
				state.CurrentShufflingSeed = state.GenerateSeed(state.CurrentShufflingEpoch)
			}
		}
	}
}

func UpdateValidatorRegistry(state *beacon.BeaconState) {
	// TODO
}
