package transition

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

// Transition the state to the given slot.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func StateTransitionTo(state *BeaconState, slot Slot) error {
	if state.Slot > slot {
		return errors.New("cannot handle block on top of pre-state with equal or higher slot than block")
	}
	// happens at the start of every Slot
	for ; state.Slot < slot; {
		CacheState(state)
		// Per-epoch transition happens at the start of the first slot of every epoch.
		if (state.Slot+1)%SLOTS_PER_EPOCH == 0 {
			EpochTransition(state)
		}
		state.Slot++
	}
	return nil
}

// Transition the state to the slot of the given block, then processes the block.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func StateTransition(state *BeaconState, block *BeaconBlock, verifyStateRoot bool) error {
	if err := StateTransitionTo(state, block.Slot); err != nil {
		return err
	}

	if err := ProcessBlock(state, block); err != nil {
		return err
	}
	// State root verification
	if verifyStateRoot && block.StateRoot != ssz.HashTreeRoot(state, BeaconStateSSZ) {
		return errors.New("block has invalid state root")
	}
	return nil
}
