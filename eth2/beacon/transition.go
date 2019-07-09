package beacon

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/block"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	"github.com/protolambda/zrnt/eth2/beacon/epoch"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessSlot(state *BeaconState) {
	// Cache latest known state root (for previous slot)
	latestStateRoot := ssz.HashTreeRoot(state, BeaconStateSSZ)
	state.StateRoots[state.Slot%SLOTS_PER_HISTORICAL_ROOT] = latestStateRoot

	// Store latest known state root (for previous slot) in latest_block_header if it is empty
	if state.LatestBlockHeader.StateRoot == (Root{}) {
		state.LatestBlockHeader.StateRoot = latestStateRoot
	}

	// Cache latest known block root (for previous slot)
	previousBlockRoot := ssz.SigningRoot(state.LatestBlockHeader, BeaconBlockHeaderSSZ)
	state.BlockRoots[state.Slot%SLOTS_PER_HISTORICAL_ROOT] = previousBlockRoot
}

// Transition the state to the given slot.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func StateTransitionTo(state *BeaconState, slot Slot) error {
	if state.Slot > slot {
		return errors.New("cannot transition from pre-state with higher slot than transition target")
	}
	// happens at the start of every Slot
	for state.Slot < slot {
		ProcessSlot(state)
		// Per-epoch transition happens at the start of the first slot of every epoch.
		if (state.Slot+1)%SLOTS_PER_EPOCH == 0 {
			epoch.Transition(state)
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

	if err := block.Transition(state); err != nil {
		return err
	}
	// State root verification
	if verifyStateRoot && block.StateRoot != ssz.HashTreeRoot(state, BeaconStateSSZ) {
		return errors.New("block has invalid state root")
	}
	return nil
}
