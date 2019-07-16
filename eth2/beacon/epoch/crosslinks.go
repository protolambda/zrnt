package epoch

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
)

func ProcessEpochCrosslinks(state *BeaconState) {
	state.PreviousCrosslinks = state.CurrentCrosslinks
	currentEpoch := state.Epoch()
	previousEpoch := state.PreviousEpoch()
	for epoch := previousEpoch; epoch <= currentEpoch; epoch++ {
		count := state.Validators.GetCommitteeCount(epoch)
		startShard := state.GetStartShard(epoch)
		for offset := uint64(0); offset < count; offset++ {
			shard := (startShard + Shard(offset)) % SHARD_COUNT
			crosslinkCommittee := state.PrecomputedData.GetCrosslinkCommittee(epoch, shard)
			winningCrosslink, attestingIndices := state.PrecomputedData.GetWinningCrosslinkAndAttesters(epoch, shard)
			participatingBalance := state.Validators.GetTotalEffectiveBalanceOf(attestingIndices)
			totalBalance := state.Validators.GetTotalEffectiveBalanceOf(crosslinkCommittee)
			if 3*participatingBalance >= 2*totalBalance {
				state.CurrentCrosslinks[shard] = *winningCrosslink
			}
		}
	}
}
