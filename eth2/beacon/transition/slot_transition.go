package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func AdvanceSlot(state *beacon.BeaconState) {
	state.Latest_state_roots[state.Slot%beacon.SLOTS_PER_HISTORICAL_ROOT] = ssz.Hash_tree_root(state)
	state.Slot += 1
	if state.LatestBlockHeader.State_root == (beacon.Root{}) {
		// previous slot is safe, ignore error
		stRoot, _ := state.Get_state_root(state.Slot-1)
		state.LatestBlockHeader.State_root = stRoot
	}
	state.Latest_block_roots[(state.Slot-1)%beacon.SLOTS_PER_HISTORICAL_ROOT] = ssz.Hash_tree_root(state.LatestBlockHeader)
}
