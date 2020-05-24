package transition

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
)

type BlockInput interface {
	Slot() Slot
	Process() error
	VerifySignature(pubkey BLSPubkey, version Version, genValRoot Root) bool
	VerifyStateRoot(expected Root) bool
}

type TransitionFeature struct {
	Meta interface {
		StartEpoch()
		CurrentSlot() Slot
		IncrementSlot()
		ProcessSlot()
		ProcessEpoch()
		StateRoot() Root
		CurrentProposer() BLSPubkey
		CurrentVersion() Version
		GenesisValRoot() Root
	}
}

// Process the state to the given slot.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func (f *TransitionFeature) ProcessSlots(slot Slot) error {
	if f.Meta.CurrentSlot() >= slot {
		return errors.New("cannot transition from pre-state with higher or equal slot than transition target")
	}
	// happens at the start of every CurrentSlot
	for f.Meta.CurrentSlot() < slot {
		f.Meta.ProcessSlot()
		// Per-epoch transition happens at the start of the first slot of every epoch.
		// (with the slot still at the end of the last epoch)
		currentSlot := f.Meta.CurrentSlot()
		isEpochEnd := (currentSlot + 1).ToEpoch() != currentSlot.ToEpoch()
		if isEpochEnd {
			f.Meta.ProcessEpoch()
		}
		f.Meta.IncrementSlot()
		if isEpochEnd {
			f.Meta.StartEpoch()
		}
	}
	return nil
}

// Transition the state to the slot of the given block, then processes the block.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
//
func (f *TransitionFeature) StateTransition(block BlockInput, validateResult bool) error {
	if err := f.ProcessSlots(block.Slot()); err != nil {
		return err
	}
	if validateResult {
		if !block.VerifySignature(f.Meta.CurrentProposer(), f.Meta.CurrentVersion(), f.Meta.GenesisValRoot()) {
			return errors.New("block has invalid signature")
		}
	}

	if err := block.Process(); err != nil {
		return err
	}

	// State root verification
	if !block.VerifyStateRoot(f.Meta.StateRoot()) {
		return errors.New("block has invalid state root")
	}
	return nil
}
