package deltas_computation

import (
	"github.com/protolambda/zrnt/eth2/beacon"
)

func DeltasCrosslinks(state *beacon.BeaconState, v beacon.Valuator) *beacon.Deltas {
	deltas := beacon.NewDeltas(uint64(len(state.ValidatorRegistry)))

	// From previous epoch start, to current epoch start
	start := state.PreviousEpoch().GetStartSlot()
	end := state.Epoch().GetStartSlot()
	for slot := start; slot < end; slot++ {
		for _, shardCommittee := range state.GetCrosslinkCommitteesAtSlot(slot, false) {
			_, participants := state.GetWinningRootAndParticipants(shardCommittee.Shard)
			participatingBalance := state.ValidatorBalances.GetTotalBalance(participants)
			totalBalance := state.ValidatorBalances.GetTotalBalance(shardCommittee.Committee)
			in, out := FindInAndOutValidators(shardCommittee.Committee, participants)
			for _, i := range in {
				deltas.Rewards[i] = v.GetBaseReward(i) * participatingBalance / totalBalance
			}
			for _, i := range out {
				deltas.Rewards[i] = v.GetBaseReward(i)
			}
		}
	}
	return deltas
}
