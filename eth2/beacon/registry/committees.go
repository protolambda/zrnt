package registry

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
)

// Return the number of committees in one epoch.
func (vr ValidatorRegistry) GetCommitteeCount(epoch Epoch) uint64 {
	activeValidatorCount := vr.GetActiveValidatorCount(epoch)
	committeesPerSlot := math.MaxU64(1,
		math.MinU64(
			uint64(SHARD_COUNT)/uint64(SLOTS_PER_EPOCH),
			activeValidatorCount/uint64(SLOTS_PER_EPOCH)/TARGET_COMMITTEE_SIZE,
		))
	return committeesPerSlot * uint64(SLOTS_PER_EPOCH)
}

// Return the number of shards to increment state.latest_start_shard during epoch
func (vr ValidatorRegistry) GetShardDelta(epoch Epoch) Shard {
	return Shard(math.MinU64(
		vr.GetCommitteeCount(epoch),
		uint64(SHARD_COUNT-(SHARD_COUNT/Shard(SLOTS_PER_EPOCH)))))
}
