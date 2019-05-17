package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
)

func ProcessEpochCrosslinks(state *BeaconState) {
	state.PreviousCrosslinks = state.CurrentCrosslinks
	currentEpoch := state.Epoch()
	previousEpoch := state.PreviousEpoch()
	nextEpoch := currentEpoch + 1
	for epoch := previousEpoch; epoch < nextEpoch; epoch++ {
		count := state.GetEpochCommitteeCount(epoch)
		startShard := state.GetEpochStartShard(epoch)
		for offset := uint64(0); offset < count; offset++ {
			shard := (startShard + Shard(offset)) % SHARD_COUNT
			crosslinkCommittee := state.GetCrosslinkCommittee(epoch, shard)
			winningCrosslink, attestingIndices := state.GetWinningCrosslinkAndAttestingIndices(shardCommittee.Shard, slot.ToEpoch())
			participatingBalance := state.GetTotalBalanceOf(attestingIndices)
			totalBalance := state.GetTotalBalanceOf(crosslinkCommittee)
			if 3*participatingBalance >= 2*totalBalance {
				state.CurrentCrosslinks[shard] = winningCrosslink
			}
		}
	}
}
