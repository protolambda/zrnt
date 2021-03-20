package beacon

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/tree"
)

// Process the state to the given slot.
// Returns an error if the slot is older than the state is already at.
// Mutates the state, does not copy.
func ProcessSlots(ctx context.Context, spec *common.Spec, epc *phase0.EpochsContext, state *phase0.BeaconStateView, slot common.Slot) error {
	// happens at the start of every CurrentSlot
	currentSlot, err := state.Slot()
	if err != nil {
		return err
	}
	if currentSlot >= slot {
		return errors.New("cannot transition from pre-state with higher or equal slot than transition target")
	}
	for currentSlot < slot {
		select {
		case <-ctx.Done():
			return common.TransitionCancelErr
		default:
			break // Continue slot processing, don't block.
		}
		if err := common.ProcessSlot(ctx, spec, state); err != nil {
			return err
		}
		// Per-epoch transition happens at the start of the first slot of every epoch.
		// (with the slot still at the end of the last epoch)
		isEpochEnd := spec.SlotToEpoch(currentSlot+1) != spec.SlotToEpoch(currentSlot)
		if isEpochEnd {
			if err := phase0.ProcessEpoch(ctx, spec, epc, state); err != nil {
				return err
			}
		}
		currentSlot += 1
		if err := state.SetSlot(currentSlot); err != nil {
			return err
		}
		if isEpochEnd {
			if err := epc.RotateEpochs(state); err != nil {
				return err
			}
		}
	}
	return nil
}

// StateTransition to the slot of the given block, then process the block.
// Returns an error if the slot is older or equal to what the state is already at.
// Mutates the state, does not copy.
func StateTransition(ctx context.Context, spec *common.Spec, epc *phase0.EpochsContext, state *phase0.BeaconStateView, block *phase0.SignedBeaconBlock, validateResult bool) error {
	if err := ProcessSlots(ctx, spec, epc, state, block.Message.Slot); err != nil {
		return err
	}
	return PostSlotTransition(ctx, spec, epc, state, block, validateResult)
}

// PostSlotTransition finishes a state transition after applying ProcessSlots(..., block.Slot).
func PostSlotTransition(ctx context.Context, spec *common.Spec, epc *phase0.EpochsContext, state *phase0.BeaconStateView, block *phase0.SignedBeaconBlock, validateResult bool) error {
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	if slot != block.Message.Slot {
		return fmt.Errorf("transition of block, post-slot-processing, must run on state with same slot")
	}
	if validateResult {
		// Safe to ignore proposer index, it will be checked as part of the ProcessHeader call.
		if !phase0.VerifyBlockSignature(spec, epc, state, block, false) {
			return errors.New("block has invalid signature")
		}
	}
	if err := phase0.ProcessBlock(ctx, spec, epc, state, &block.Message); err != nil {
		return err
	}

	// State root verification
	if validateResult && block.Message.StateRoot != state.HashTreeRoot(tree.GetHashFn()) {
		return errors.New("block has invalid state root")
	}
	return nil
}
