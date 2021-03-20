package phase0

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func ProcessEpochJustification(ctx context.Context, spec *common.Spec, epc *EpochsContext, process *EpochProcess, state common.BeaconState) error {
	select {
	case <-ctx.Done():
		return common.TransitionCancelErr
	default: // Don't block.
		break
	}
	previousEpoch := process.PrevEpoch
	currentEpoch := process.CurrEpoch

	// skip if genesis.
	if currentEpoch <= common.GENESIS_EPOCH+1 {
		return nil
	}

	oldPreviousJustified, err := state.PreviousJustifiedCheckpoint()
	if err != nil {
		return err
	}
	oldCurrentJustified, err := state.CurrentJustifiedCheckpoint()
	if err != nil {
		return err
	}

	bits, err := state.JustificationBits()
	if err != nil {
		return err
	}

	// Rotate (a copy of) current into previous
	if err := state.SetPreviousJustifiedCheckpoint(oldCurrentJustified); err != nil {
		return err
	}

	bits.NextEpoch()

	// Get the total current stake
	totalStake := process.TotalActiveStake

	var newJustifiedCheckpoint *common.Checkpoint
	// > Justification
	if process.PrevEpochUnslashedStake.TargetStake*3 >= totalStake*2 {
		root, err := common.GetBlockRoot(spec, state, previousEpoch)
		if err != nil {
			return err
		}
		newJustifiedCheckpoint = &common.Checkpoint{
			Epoch: previousEpoch,
			Root:  root,
		}
		bits[0] |= 1 << 1
	}
	if process.CurrEpochUnslashedTargetStake*3 >= totalStake*2 {
		root, err := common.GetBlockRoot(spec, state, currentEpoch)
		if err != nil {
			return err
		}
		newJustifiedCheckpoint = &common.Checkpoint{
			Epoch: currentEpoch,
			Root:  root,
		}
		bits[0] |= 1 << 0
	}
	if newJustifiedCheckpoint != nil {
		if err := state.SetCurrentJustifiedCheckpoint(*newJustifiedCheckpoint); err != nil {
			return err
		}
	}

	// > Finalization
	var toFinalize *common.Checkpoint
	// The 2nd/3rd/4th most recent epochs are all justified, the 2nd using the 4th as source
	if justified := bits.IsJustified(1, 2, 3); justified && oldPreviousJustified.Epoch+3 == currentEpoch {
		toFinalize = &oldPreviousJustified
	}
	// The 2nd/3rd most recent epochs are both justified, the 2nd using the 3rd as source
	if justified := bits.IsJustified(1, 2); justified && oldPreviousJustified.Epoch+2 == currentEpoch {
		toFinalize = &oldPreviousJustified
	}
	// The 1st/2nd/3rd most recent epochs are all justified, the 1st using the 3rd as source
	if justified := bits.IsJustified(0, 1, 2); justified && oldCurrentJustified.Epoch+2 == currentEpoch {
		toFinalize = &oldCurrentJustified
	}
	// The 1st/2nd most recent epochs are both justified, the 1st using the 2nd as source
	if justified := bits.IsJustified(0, 1); justified && oldCurrentJustified.Epoch+1 == currentEpoch {
		toFinalize = &oldCurrentJustified
	}
	if toFinalize != nil {
		if err := state.SetFinalizedCheckpoint(*toFinalize); err != nil {
			return err
		}
	}
	if err := state.SetJustificationBits(bits); err != nil {
		return err
	}
	return nil
}
