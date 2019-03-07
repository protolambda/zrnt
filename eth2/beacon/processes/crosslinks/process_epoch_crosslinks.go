package crosslinks

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

func ProcessEpochCrosslinks(state *beacon.BeaconState) {
	current_epoch := state.Epoch()
	previous_epoch := current_epoch - 1
	next_epoch := current_epoch + 1
	start := previous_epoch.GetStartSlot()
	end := next_epoch.GetStartSlot()
	for slot := start; slot < end; slot++ {
		for _, shard_committee := range state.Get_crosslink_committees_at_slot(slot, false) {
			winning_root, participants := state.Get_winning_root_and_participants(shard_committee.Shard)
			participating_balance := state.Validator_balances.Get_total_balance(participants)
			total_balance := state.Validator_balances.Get_total_balance(shard_committee.Committee)
			if 3*participating_balance >= 2*total_balance {
				state.Latest_crosslinks[shard_committee.Shard] = beacon.Crosslink{
					Epoch:               slot.ToEpoch(),
					Crosslink_data_root: winning_root,
				}
			}
		}
	}
}
