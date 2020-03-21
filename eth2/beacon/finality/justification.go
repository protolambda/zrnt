package finality

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type JustificationEpochProcess interface {
	ProcessEpochJustification(input JustificationEpochProcessInput) error
}

type JustificationEpochProcessInput interface {
	meta.Versioning
	meta.History
	meta.Staking
}

func (state *FinalityProps) ProcessEpochJustification(input JustificationEpochProcessInput) error {
	currentEpoch, err := input.CurrentEpoch()
	if err != nil {
		return err
	}
	// skip if genesis.
	if currentEpoch <= GENESIS_EPOCH+1 {
		return nil
	}
	previousEpoch, err := input.PreviousEpoch()
	if err != nil {
		return err
	}

	// stake = effective balances of active validators
	// Get the total stake of the epoch attesters
	prevEpochStake, err := input.PrevEpochStakeSummary()
	if err != nil {
		return err
	}
	currEpochStake, err := input.CurrEpochStakeSummary()
	if err != nil {
		return err
	}

	// Get the total current stake
	totalStake, err := input.GetTotalStake()
	if err != nil {
		return err
	}

	oldPreviousJustified, err := state.PreviousJustifiedCheckpoint.CheckPoint()
	if err != nil {
		return err
	}
	oldCurrentJustified, err := state.CurrentJustifiedCheckpoint.CheckPoint()
	if err != nil {
		return err
	}

	// Rotate current into previous
	if err := state.PreviousJustifiedCheckpoint.SetCheckPoint(oldCurrentJustified); err != nil {
		return err
	}
	if err := state.JustificationBits.NextEpoch(); err != nil {
		return err
	}

	// > Justification
	if prevEpochStake.TargetStake*3 >= totalStake*2 {
		root, err := input.GetBlockRoot(previousEpoch)
		if err != nil {
			return err
		}
		if err := state.Justify(currentEpoch, Checkpoint{
			Epoch: previousEpoch,
			Root:  root,
		}); err != nil {
			return err
		}
	}
	if currEpochStake.TargetStake*3 >= totalStake*2 {
		root, err := input.GetBlockRoot(currentEpoch)
		if err != nil {
			return err
		}
		if err := state.Justify(currentEpoch, Checkpoint{
			Epoch: currentEpoch,
			Root:  root,
		}); err != nil {
			return err
		}
	}

	// > Finalization
	bits := state.JustificationBits
	var toFinalize *Checkpoint
	// The 2nd/3rd/4th most recent epochs are all justified, the 2nd using the 4th as source
	if justified, err := bits.IsJustified(1, 2, 3); err != nil {
		return err
	} else if justified && oldPreviousJustified.Epoch+3 == currentEpoch {
		toFinalize = &oldPreviousJustified
	}
	// The 2nd/3rd most recent epochs are both justified, the 2nd using the 3rd as source
	if justified, err := bits.IsJustified(1, 2); err != nil {
		return err
	} else if justified && oldPreviousJustified.Epoch+2 == currentEpoch {
		toFinalize = &oldPreviousJustified
	}
	// The 1st/2nd/3rd most recent epochs are all justified, the 1st using the 3rd as source
	if justified, err := bits.IsJustified(0, 1, 2); err != nil {
		return err
	} else if justified && oldCurrentJustified.Epoch+2 == currentEpoch {
		toFinalize = &oldCurrentJustified
	}
	// The 1st/2nd most recent epochs are both justified, the 1st using the 2nd as source
	if justified, err := bits.IsJustified(0, 1); err != nil {
		return err
	} else if justified && oldCurrentJustified.Epoch+1 == currentEpoch {
		toFinalize = &oldCurrentJustified
	}
	if toFinalize != nil {
		if err := state.FinalizedCheckpoint.SetCheckPoint(*toFinalize); err != nil {
			return err
		}
	}
	return nil
}
