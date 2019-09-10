package finality

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type JustificationEpochProcess interface {
	ProcessEpochJustification()
}

func (f *JustificationFeature) ProcessEpochJustification() {
	currentEpoch := f.Meta.CurrentEpoch()
	if currentEpoch <= GENESIS_EPOCH+1 {
		return
	}
	previousEpoch := f.Meta.PreviousEpoch()

	// stake = effective balances of active validators
	// Get the total stake of the epoch attesters
	attesterStatuses := f.Meta.GetAttesterStatuses()
	prevTargetStake := f.Meta.GetAttestersStake(attesterStatuses, PrevTargetAttester|UnslashedAttester)
	currTargetStake := f.Meta.GetAttestersStake(attesterStatuses, CurrTargetAttester|UnslashedAttester)

	// Get the total current stake
	totalStake := f.Meta.GetTotalStake()

	oldPreviousJustified := f.State.PreviousJustifiedCheckpoint
	oldCurrentJustified := f.State.CurrentJustifiedCheckpoint

	// Rotate current into previous
	f.State.PreviousJustifiedCheckpoint = f.State.CurrentJustifiedCheckpoint
	f.State.JustificationBits.NextEpoch()

	// > Justification
	if prevTargetStake*3 >= totalStake*2 {
		f.Justify(Checkpoint{
			Epoch: previousEpoch,
			Root:  f.Meta.GetBlockRoot(previousEpoch),
		})
	}
	if currTargetStake*3 >= totalStake*2 {
		f.Justify(Checkpoint{
			Epoch: currentEpoch,
			Root:  f.Meta.GetBlockRoot(currentEpoch),
		})
	}

	// > Finalization
	bits := f.State.JustificationBits
	// The 2nd/3rd/4th most recent epochs are all justified, the 2nd using the 4th as source
	if bits.IsJustified(1, 2, 3) && oldPreviousJustified.Epoch+3 == currentEpoch {
		f.State.FinalizedCheckpoint = oldPreviousJustified
	}
	// The 2nd/3rd most recent epochs are both justified, the 2nd using the 3rd as source
	if bits.IsJustified(1, 2) && oldPreviousJustified.Epoch+2 == currentEpoch {
		f.State.FinalizedCheckpoint = oldPreviousJustified
	}
	// The 1st/2nd/3rd most recent epochs are all justified, the 1st using the 3rd as source
	if bits.IsJustified(0, 1, 2) && oldCurrentJustified.Epoch+2 == currentEpoch {
		f.State.FinalizedCheckpoint = oldCurrentJustified
	}
	// The 1st/2nd most recent epochs are both justified, the 1st using the 2nd as source
	if bits.IsJustified(0, 1) && oldCurrentJustified.Epoch+1 == currentEpoch {
		f.State.FinalizedCheckpoint = oldCurrentJustified
	}
}
