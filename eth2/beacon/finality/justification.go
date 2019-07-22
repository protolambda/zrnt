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

	// epoch numbers are trusted, no errors
	previousBoundaryBlockRoot := f.Meta.GetBlockRoot(previousEpoch)
	currentBoundaryBlockRoot := f.Meta.GetBlockRoot(currentEpoch)

	// Get the sum balances of the boundary attesters
	previousTargetStakedBalance := f.Meta.GetTargetTotalStakedBalance(previousEpoch)
	currentTargetStakedBalance := f.Meta.GetTargetTotalStakedBalance(currentEpoch)

	// Get the sum balances of the attesters for the epochs
	previousTotalBalance := f.Meta.GetTotalStakedBalance(previousEpoch)
	currentTotalBalance := f.Meta.GetTotalStakedBalance(currentEpoch)

	oldPreviousJustified := f.State.PreviousJustifiedCheckpoint
	oldCurrentJustified := f.State.CurrentJustifiedCheckpoint

	// Rotate current into previous
	f.State.PreviousJustifiedCheckpoint = f.State.CurrentJustifiedCheckpoint
	f.State.JustificationBits.NextEpoch()

	// > Justification
	if previousTargetStakedBalance*3 >= previousTotalBalance*2 {
		f.Justify(Checkpoint{
			Epoch: previousEpoch,
			Root:  previousBoundaryBlockRoot,
		})
	}
	if currentTargetStakedBalance*3 >= currentTotalBalance*2 {
		f.Justify(Checkpoint{
			Epoch: currentEpoch,
			Root:  currentBoundaryBlockRoot,
		})
	}

	// > Finalization
	bits := f.State.JustificationBits
	// The 2nd/3rd/4th most recent epochs are all justified, the 2nd using the 4th as source
	if bits.IsJustified(1, 2, 3) && f.State.PreviousJustifiedCheckpoint.Epoch+3 == currentEpoch {
		f.State.FinalizedCheckpoint = oldPreviousJustified
	}
	// The 2nd/3rd most recent epochs are both justified, the 2nd using the 3rd as source
	if bits.IsJustified(1, 2) && f.State.PreviousJustifiedCheckpoint.Epoch+2 == currentEpoch {
		f.State.FinalizedCheckpoint = oldPreviousJustified
	}
	// The 1st/2nd/3rd most recent epochs are all justified, the 1st using the 3rd as source
	if bits.IsJustified(0, 1, 2) && f.State.CurrentJustifiedCheckpoint.Epoch+2 == currentEpoch {
		f.State.FinalizedCheckpoint = oldCurrentJustified
	}
	// The 1st/2nd most recent epochs are both justified, the 1st using the 2nd as source
	if bits.IsJustified(0, 1) && f.State.CurrentJustifiedCheckpoint.Epoch+1 == currentEpoch {
		f.State.FinalizedCheckpoint = oldCurrentJustified
	}
}
