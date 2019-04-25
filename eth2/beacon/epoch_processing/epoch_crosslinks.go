package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
)

func ProcessEpochCrosslinks(state *BeaconState) {
	state.PreviousCrosslinks = state.CurrentCrosslinks
	currentEpoch := state.Epoch()
	previousEpoch := state.PreviousEpoch()
	nextEpoch := currentEpoch + 1
	start := previousEpoch.GetStartSlot()
	end := nextEpoch.GetStartSlot()
	for slot := start; slot < end; slot++ {
		for _, shardCommittee := range state.GetCrosslinkCommitteesAtSlot(slot) {
			winningCrosslink, attestingIndices := state.GetWinningCrosslinkAndAttestingIndices(shardCommittee.Shard, slot.ToEpoch())
			participatingBalance := state.GetTotalBalanceOf(attestingIndices)
			totalBalance := state.GetTotalBalanceOf(shardCommittee.Committee)
			if 3*participatingBalance >= 2*totalBalance {
				state.CurrentCrosslinks[shardCommittee.Shard] = winningCrosslink
			}
		}
	}
}
