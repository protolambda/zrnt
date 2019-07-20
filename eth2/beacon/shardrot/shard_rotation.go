package shardrot

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/math"
)

type ShardRotationState struct {
	StartShard Shard
}

// Return the number of shards to increment state.StartShard during epoch
func (state *ShardRotationState) GetShardDelta(meta CommitteeCountMeta, epoch Epoch) Shard {
	return Shard(math.MinU64(
		meta.GetCommitteeCount(epoch),
		uint64(SHARD_COUNT-(SHARD_COUNT/Shard(SLOTS_PER_EPOCH)))))
}

type RotShardReq interface {
	VersioningMeta
	CommitteeCountMeta
}

func (state *ShardRotationState) RotateStartShard(meta RotShardReq) {
	state.StartShard = (state.StartShard + state.GetShardDelta(meta, meta.Epoch())) % SHARD_COUNT
}
