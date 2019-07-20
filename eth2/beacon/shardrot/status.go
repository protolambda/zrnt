package shardrot

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
)

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

type StartShardsReq interface {
	VersioningMeta
	CommitteeCountMeta
	CrosslinkTimingMeta
}

// Load start shards, starting from fromEpoch, until the next epoch (incl.)
func (state *ShardRotationState) LoadStartShardStatus(meta StartShardsReq, fromEpoch Epoch) *StartShardStatus {
	res := new(StartShardStatus)
	currentEpoch := meta.Epoch()
	res.LatestStartEpoch = currentEpoch + 1
	shard := (state.StartShard + state.GetShardDelta(meta, currentEpoch)) % SHARD_COUNT
	res.StartShards = append(res.StartShards, shard)
	diff := currentEpoch
	if fromEpoch < currentEpoch {
		diff = currentEpoch - fromEpoch
	}
	for i := Epoch(0); i <= diff; i++ {
		epoch := currentEpoch - i
		shard = (shard + SHARD_COUNT - state.GetShardDelta(meta, epoch)) % SHARD_COUNT
		res.StartShards = append(res.StartShards, shard)
	}
	return res
}
