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

type SlotsProcessInput interface {
	CurrentSlot() Slot
	IncrementSlot()
	ProcessSlot()
	ProcessEpoch()
}

// Process the state to the given slot.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func ProcessSlots(input SlotsProcessInput, slot Slot) {
	// happens at the start of every CurrentSlot
	for input.CurrentSlot() < slot {
		input.ProcessSlot()
		// Per-epoch transition happens at the start of the first slot of every epoch.
		// (with the slot still at the end of the last epoch)
		currentSlot := input.CurrentSlot()
		isEpochEnd := (currentSlot + 1).ToEpoch() != currentSlot.ToEpoch()
		if isEpochEnd {
			input.ProcessEpoch()
		}
		input.IncrementSlot()
	}
}

type TransitionProcessInput interface {
	SlotsProcessInput
	StateRoot() Root
	CurrentProposer() BLSPubkey
	CurrentVersion() Version
	GenesisValRoot() Root
}

// Transition the state to the slot of the given block, then processes the block.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
//
func StateTransition(input TransitionProcessInput, block BlockInput, validateResult bool) error {
	if input.CurrentSlot() > block.Slot() {
		return errors.New("cannot transition from pre-state with higher slot than transition target")
	}
	ProcessSlots(input, block.Slot())
	if validateResult {
		if !block.VerifySignature(input.CurrentProposer(), input.CurrentVersion(), input.GenesisValRoot()) {
			return errors.New("block has invalid signature")
		}
	}

	if err := block.Process(); err != nil {
		return err
	}

	// State root verification
	if validateResult && !block.VerifyStateRoot(input.StateRoot()) {
		return errors.New("block has invalid state root")
	}
	return nil
}
