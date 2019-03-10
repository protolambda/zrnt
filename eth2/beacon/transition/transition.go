package transition

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func StateTransition(preState *beacon.BeaconState, block *beacon.BeaconBlock, verifyStateRoot bool) (res *beacon.BeaconState, err error) {
	if preState.Slot >= block.Slot {
		return nil, errors.New("cannot handle block on top of pre-state with equal or higher slot than block")
	}
	// We work on a copy of the input state. If the block is invalid, or input is re-used, we don't have to care.
	state := preState.Copy()
	// happens at the start of every Slot
	for i := state.Slot; i < block.Slot; i++ {
		AdvanceSlot(state)
	}
	// happens at every block
	if err := ApplyBlock(state, block); err != nil {
		return nil, err
	}
	// "happens at the end of the last Slot of every epoch "
	if (state.Slot+1)%beacon.SLOTS_PER_EPOCH == 0 {
		EpochTransition(state)
	}
	// State root verification
	if verifyStateRoot && block.StateRoot != ssz.HashTreeRoot(state) {
		return nil, errors.New("block has invalid state root")
	}
	return state, nil
}
