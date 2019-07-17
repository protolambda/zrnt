package components

import . "github.com/protolambda/zrnt/eth2/core"

type ShardRotationState struct {
	StartShard Shard
}

type StartShardReq interface {
	VersioningMeta
	ShardDeltaMeta
}

func (state *ShardRotationState) GetStartShard(meta StartShardReq, epoch Epoch) Shard {
	currentEpoch := meta.Epoch()
	checkEpoch := currentEpoch + 1
	if epoch > checkEpoch {
		panic("cannot find start shard for epoch, epoch is too new")
	}
	shard := (state.StartShard + meta.GetShardDelta(currentEpoch)) % SHARD_COUNT
	for checkEpoch > epoch {
		checkEpoch--
		shard = (shard + SHARD_COUNT - meta.GetShardDelta(checkEpoch)) % SHARD_COUNT
	}
	return shard
}

func (state *ShardRotationState) RotateStartShard(meta StartShardReq) {
	state.StartShard = (state.StartShard + meta.GetShardDelta(meta.Epoch())) % SHARD_COUNT
}
