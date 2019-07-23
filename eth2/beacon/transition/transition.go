package transition

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
)

type BlockInput interface {
	Slot() Slot
	Process() error
	StateRoot() Root
}

type TransitionFeature struct {
	Meta interface {
		CurrentSlot() Slot
		IncrementSlot()
		ProcessSlot()
		ProcessEpoch()
		StateRoot() Root
	}
}

// Process the state to the given slot.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func (f *TransitionFeature) ProcessSlots(slot Slot) {
	// happens at the start of every CurrentSlot
	for f.Meta.CurrentSlot() < slot {
		f.Meta.ProcessSlot()
		// Per-epoch transition happens at the start of the first slot of every epoch.
		// (with the slot still at the end of the last epoch)
		if slot := f.Meta.CurrentSlot(); (slot+1).ToEpoch() != slot.ToEpoch() {
			f.Meta.ProcessEpoch()
		}
		f.Meta.IncrementSlot()
	}
}

// Transition the state to the slot of the given block, then processes the block.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func (f *TransitionFeature) StateTransition(block BlockInput, verifyStateRoot bool) error {
	if f.Meta.CurrentSlot() > block.Slot() {
		return errors.New("cannot transition from pre-state with higher slot than transition target")
	}
	f.ProcessSlots(block.Slot())

	if err := block.Process(); err != nil {
		return err
	}

	// State root verification
	if verifyStateRoot && block.StateRoot() != f.Meta.StateRoot() {
		return errors.New("block has invalid state root")
	}
	return nil
}
