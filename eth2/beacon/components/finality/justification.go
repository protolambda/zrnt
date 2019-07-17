package finality

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/meta"
	. "github.com/protolambda/zrnt/eth2/core"
)

type JustificationReq interface {
	VersioningMeta
	HistoryMeta
	FFGMeta
}

func (state *FinalityState) ProcessEpochJustification(meta JustificationReq) {
	currentEpoch := meta.Epoch()
	if currentEpoch <= GENESIS_EPOCH+1 {
		return
	}
	previousEpoch := meta.PreviousEpoch()

	// epoch numbers are trusted, no errors
	previousBoundaryBlockRoot := meta.GetBlockRoot(previousEpoch)
	currentBoundaryBlockRoot := meta.GetBlockRoot(currentEpoch)

	// Get the sum balances of the boundary attesters
	previousTargetStakedBalance := meta.GetTargetTotalStakedBalance(previousEpoch)
	currentTargetStakedBalance := meta.GetTargetTotalStakedBalance(currentEpoch)

	// Get the sum balances of the attesters for the epochs
	previousTotalBalance := meta.GetTotalStakedBalance(previousEpoch)
	currentTotalBalance := meta.GetTotalStakedBalance(currentEpoch)

	oldPreviousJustified := state.PreviousJustifiedCheckpoint
	oldCurrentJustified := state.CurrentJustifiedCheckpoint

	// Rotate current into previous
	state.PreviousJustifiedCheckpoint = state.CurrentJustifiedCheckpoint
	state.JustificationBits.NextEpoch()

	// > Justification
	if previousTargetStakedBalance*3 >= previousTotalBalance*2 {
		state.Justify(meta, Checkpoint{
			Epoch: previousEpoch,
			Root:  previousBoundaryBlockRoot,
		})
	}
	if currentTargetStakedBalance*3 >= currentTotalBalance*2 {
		state.Justify(meta, Checkpoint{
			Epoch: currentEpoch,
			Root:  currentBoundaryBlockRoot,
		})
	}

	// > Finalization
	bits := state.JustificationBits
	// The 2nd/3rd/4th most recent epochs are all justified, the 2nd using the 4th as source
	if bits.IsJustified(1, 2, 3) && state.PreviousJustifiedCheckpoint.Epoch+3 == currentEpoch {
		state.FinalizedCheckpoint = oldPreviousJustified
	}
	// The 2nd/3rd most recent epochs are both justified, the 2nd using the 3rd as source
	if bits.IsJustified(1, 2) && state.PreviousJustifiedCheckpoint.Epoch+2 == currentEpoch {
		state.FinalizedCheckpoint = oldPreviousJustified
	}
	// The 1st/2nd/3rd most recent epochs are all justified, the 1st using the 3rd as source
	if bits.IsJustified(0, 1, 2) && state.CurrentJustifiedCheckpoint.Epoch+2 == currentEpoch {
		state.FinalizedCheckpoint = oldCurrentJustified
	}
	// The 1st/2nd most recent epochs are both justified, the 1st using the 2nd as source
	if bits.IsJustified(0, 1) && state.CurrentJustifiedCheckpoint.Epoch+1 == currentEpoch {
		state.FinalizedCheckpoint = oldCurrentJustified
	}
}
