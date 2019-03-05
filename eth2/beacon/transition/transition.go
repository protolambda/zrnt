package transition

import (
	"errors"
	"go-beacon-transition/eth2"
	"go-beacon-transition/eth2/beacon"
	"go-beacon-transition/eth2/util/ssz"
)

func StateTransition(preState *beacon.BeaconState, block *beacon.BeaconBlock) (res *beacon.BeaconState, err error) {
	if preState.Slot >= block.Slot {
		return nil, errors.New("cannot handle block on top of pre-state with equal or higher slot than block")
	}
	// We work on a copy of the input state. If the block is invalid, or input is re-used, we don't have to care.
	state := preState.Copy()
	// happens at the start of every Slot
	for i := state.Slot; i < block.Slot; i++ {
		// Verified earlier, before calling StateTransition:
		// > The parent block with root `block.parent_root` has been processed and accepted
		// Hence, we can update latest block roots with the parent block root
		SlotTransition(state, block.Parent_root)
	}
	// happens at every block
	if err := ApplyBlock(state, block); err != nil {
		return nil, err
	}
	// "happens at the end of the last Slot of every epoch "
	if (state.Slot+1)%eth2.SLOTS_PER_EPOCH == 0 {
		EpochTransition(state)
	}
	// State root verification
	if block.State_root != ssz.Hash_tree_root(state) {
		return nil, errors.New("block has invalid state root")
	}
	return state, nil
}

