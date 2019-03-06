package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func AdvanceSlot(state *beacon.BeaconState) {
	state.Latest_state_roots[state.Slot % eth2.SLOTS_PER_HISTORICAL_ROOT] = ssz.Hash_tree_root(state)
	state.Slot += 1
	if state.LatestBlockHeader.State_root == (eth2.Root{}) {
		// previous slot is safe, ignore error
		stRoot, _ := Get_state_root(state, state.Slot-1)
		state.LatestBlockHeader.State_root = stRoot
	}
	state.Latest_block_roots[(state.Slot - 1) % eth2.SLOTS_PER_HISTORICAL_ROOT] = ssz.Hash_tree_root(state.latest_block_header)
}
