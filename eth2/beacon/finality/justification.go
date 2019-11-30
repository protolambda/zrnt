package finality

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type JustificationEpochProcess interface {
	ProcessEpochJustification() error
}

func (f *JustificationFeature) ProcessEpochJustification() error {
	currentEpoch, err := f.Meta.CurrentEpoch()
	if err != nil {
		return err
	}
	// skip if genesis.
	if currentEpoch <= GENESIS_EPOCH+1 {
		return nil
	}
	previousEpoch, err := f.Meta.PreviousEpoch()
	if err != nil {
		return err
	}

	// stake = effective balances of active validators
	// Get the total stake of the epoch attesters
	attesterStatuses, err := f.Meta.GetAttesterStatuses()
	if err != nil {
		return err
	}
	prevTargetStake, err := f.Meta.GetAttestersStake(attesterStatuses, PrevTargetAttester|UnslashedAttester)
	if err != nil {
		return err
	}
	currTargetStake, err := f.Meta.GetAttestersStake(attesterStatuses, CurrTargetAttester|UnslashedAttester)
	if err != nil {
		return err
	}

	// Get the total current stake
	totalStake, err := f.Meta.GetTotalStake()
	if err != nil {
		return err
	}

	oldPreviousJustified, err := f.State.PreviousJustifiedCheckpoint.CheckPoint()
	if err != nil {
		return err
	}
	oldCurrentJustified, err := f.State.CurrentJustifiedCheckpoint.CheckPoint()
	if err != nil {
		return err
	}

	// Rotate current into previous
	if err := f.State.PreviousJustifiedCheckpoint.SetCheckPoint(oldCurrentJustified); err != nil {
		return err
	}
	if err := f.State.JustificationBits.NextEpoch(); err != nil {
		return err
	}

	// > Justification
	if prevTargetStake*3 >= totalStake*2 {
		root, err := f.Meta.GetBlockRoot(previousEpoch)
		if err != nil {
			return err
		}
		if err := f.Justify(Checkpoint{
			Epoch: previousEpoch,
			Root:  root,
		}); err != nil {
			return err
		}
	}
	if currTargetStake*3 >= totalStake*2 {
		root, err := f.Meta.GetBlockRoot(currentEpoch)
		if err != nil {
			return err
		}
		if err := f.Justify(Checkpoint{
			Epoch: currentEpoch,
			Root:  root,
		}); err != nil {
			return err
		}
	}

	// > Finalization
	bits := f.State.JustificationBits
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
		if err := f.State.FinalizedCheckpoint.SetCheckPoint(*toFinalize); err != nil {
			return err
		}
	}
	return nil
}
