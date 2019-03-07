package crosslinks

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/stake"
)

func DeltasCrosslinks(state *beacon.BeaconState) *stake.Deltas {
	return stake.NewDeltas(uint64(len(state.Validator_registry)))
	// TODO: implement crosslinks rewards again
	//// Crosslinks should be created by the committees
	//start, end := previous_epoch.GetStartSlot(), next_epoch.GetStartSlot()
	//for slot := start; slot < end; slot++ {
	//	// epoch is trusted, ignore error
	//	crosslink_committees_at_slot, _ := Get_crosslink_committees_at_slot(state, slot, false)
	//	for _, cross_comm := range crosslink_committees_at_slot {
	//
	//		// We remembered the winning root
	//		// (i.e. the most attested crosslink root, doesn't have to be 2/3 majority)
	//		winning_root := winning_roots[cross_comm.Shard]
	//
	//		// We remembered the attesters of the crosslink
	//		crosslink_attesters := crosslink_winners[winning_root]
	//
	//		// Note: non-committee validators still count as attesters for a crosslink,
	//		//  hence the extra work to filter for just the validators in the committee
	//		committee_non_participants := validatorset.ValidatorIndexSet(cross_comm.Committee).Minus(crosslink_attesters)
	//
	//		committee_attesters_weight := crosslink_winners_weight[winning_root]
	//		total_committee_weight := Get_total_balance(state, cross_comm.Committee)
	//
	//		// Reward those that contributed to finding a winning root.
	//		applyRewardOrSlash(validatorset.ValidatorIndexSet(cross_comm.Committee).Minus(committee_non_participants),
	//			true, func(index beacon.ValidatorIndex) beacon.Gwei {
	//				return base_reward(index) * committee_attesters_weight / total_committee_weight
	//			})
	//		// Slash those that opted for a different crosslink
	//		applyRewardOrSlash(committee_non_participants, false, base_reward)
	//	}
	//}
}