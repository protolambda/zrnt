package shardrot

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/math"
)

type ShardRotFeature struct {
	*ShardRotationState
	Meta interface {
		VersioningMeta
		CommitteeCountMeta
	}
}

type ShardRotationState struct {
	StartShard Shard
}

// Return the number of shards to increment state.StartShard during epoch
func (state *ShardRotFeature) GetShardDelta(epoch Epoch) Shard {
	return Shard(math.MinU64(
		state.Meta.GetCommitteeCount(epoch),
		uint64(SHARD_COUNT-(SHARD_COUNT/Shard(SLOTS_PER_EPOCH)))))
}

func (state *ShardRotFeature) RotateStartShard() {
	state.StartShard = (state.StartShard + state.GetShardDelta(state.Meta.CurrentEpoch())) % SHARD_COUNT
}

type StartShardStatus struct {
	StartShards []Shard
	LatestStartEpoch Epoch
}

func (status *StartShardStatus) GetStartShard(epoch Epoch) Shard {
	if epoch > status.LatestStartEpoch {
		panic("cannot find start shard for epoch, epoch is too new")
	}
	if epoch + Epoch(len(status.StartShards)) < status.LatestStartEpoch {
		panic("cannot find start shard for epoch, epoch is too old")
	}
	return status.StartShards[status.LatestStartEpoch - epoch]
}

// Load start shards, starting from fromEpoch, until the next epoch (incl.)
func (state *ShardRotFeature) LoadStartShardStatus(fromEpoch Epoch) *StartShardStatus {
	res := new(StartShardStatus)
	currentEpoch := state.Meta.CurrentEpoch()
	res.LatestStartEpoch = currentEpoch + 1
	shard := (state.StartShard + state.GetShardDelta(currentEpoch)) % SHARD_COUNT
	res.StartShards = append(res.StartShards, shard)
	diff := currentEpoch
	if fromEpoch < currentEpoch {
		diff = currentEpoch - fromEpoch
	}
	for i := Epoch(0); i <= diff; i++ {
		epoch := currentEpoch - i
		shard = (shard + SHARD_COUNT - state.GetShardDelta(epoch)) % SHARD_COUNT
		res.StartShards = append(res.StartShards, shard)
	}
	return res
}
