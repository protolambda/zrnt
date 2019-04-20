package transition

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func StateTransition(preState *BeaconState, block *BeaconBlock, verifyStateRoot bool) (res *BeaconState, err error) {
	if preState.Slot >= block.Slot {
		return nil, errors.New("cannot handle block on top of pre-state with equal or higher slot than block")
	}
	// We work on a copy of the input state. If the block is invalid, or input is re-used, we don't have to care.
	state := preState.Copy()
	// happens at the start of every Slot
	for ; state.Slot < block.Slot; {
		CacheState(state)
		// Per-epoch transition happens at the start of the first slot of every epoch.
		if (state.Slot+1)%SLOTS_PER_EPOCH == 0 {
			EpochTransition(state)
		}
		state.Slot++
	}
	// happens at every block
	if err := ProcessBlock(state, block); err != nil {
		return nil, err
	}
	// State root verification
	if verifyStateRoot && block.StateRoot != ssz.HashTreeRoot(state) {
		return nil, errors.New("block has invalid state root")
	}
	return state, nil
}
