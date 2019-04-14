package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
)

func ProcessEpochCrosslinks(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	previousEpoch := currentEpoch - 1
	if previousEpoch < beacon.GENESIS_EPOCH {
		previousEpoch = beacon.GENESIS_EPOCH
	}
	nextEpoch := currentEpoch + 1
	start := previousEpoch.GetStartSlot()
	end := nextEpoch.GetStartSlot()
	for slot := start; slot < end; slot++ {
		for _, shardCommittee := range state.GetCrosslinkCommitteesAtSlot(slot) {
			winningRoot, participants := state.GetWinningRootAndParticipants(shardCommittee.Shard)
			participatingBalance := state.GetTotalBalanceOf(participants)
			totalBalance := state.GetTotalBalanceOf(shardCommittee.Committee)
			if 3*participatingBalance >= 2*totalBalance {
				epoch := slot.ToEpoch()
				if alt := state.LatestCrosslinks[shardCommittee.Shard].Epoch + beacon.MAX_CROSSLINK_EPOCHS; alt < epoch {
					epoch = alt
				}
				state.LatestCrosslinks[shardCommittee.Shard] = beacon.Crosslink{
					Epoch:             epoch,
					CrosslinkDataRoot: winningRoot,
				}
			}
		}
	}
}
