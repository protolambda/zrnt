package crosslinks

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

func ProcessEpochCrosslinks(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	previousEpoch := currentEpoch - 1
	nextEpoch := currentEpoch + 1
	start := previousEpoch.GetStartSlot()
	end := nextEpoch.GetStartSlot()
	for slot := start; slot < end; slot++ {
		for _, shardCommittee := range state.GetCrosslinkCommitteesAtSlot(slot, false) {
			winningRoot, participants := state.GetWinningRootAndParticipants(shardCommittee.Shard)
			participatingBalance := state.ValidatorBalances.GetTotalBalance(participants)
			totalBalance := state.ValidatorBalances.GetTotalBalance(shardCommittee.Committee)
			if 3*participatingBalance >= 2*totalBalance {
				state.LatestCrosslinks[shardCommittee.Shard] = beacon.Crosslink{
					Epoch:             slot.ToEpoch(),
					CrosslinkDataRoot: winningRoot,
				}
			}
		}
	}
}
