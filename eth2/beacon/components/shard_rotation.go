package components

import . "github.com/protolambda/zrnt/eth2/core"

type ShardRotationState struct {
	StartShard Shard
}

func (state *BeaconState) GetEpochStartShard(epoch Epoch) Shard {
	currentEpoch := state.Epoch()
	checkEpoch := currentEpoch + 1
	if epoch > checkEpoch {
		panic("cannot find start shard for epoch, epoch is too new")
	}
	shard := (state.StartShard + state.Validators.GetShardDelta(currentEpoch)) % SHARD_COUNT
	for checkEpoch > epoch {
		checkEpoch--
		shard = (shard + SHARD_COUNT - state.Validators.GetShardDelta(checkEpoch)) % SHARD_COUNT
	}
	return shard
}

func (state *BeaconState) RotateStartShard() {
	state.StartShard = (state.StartShard + state.Validators.GetShardDelta(state.Epoch())) % SHARD_COUNT
}
