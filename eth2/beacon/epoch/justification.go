package epoch

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
)

func ProcessEpochJustification(state *BeaconState) {
	currentEpoch := state.Epoch()
	if currentEpoch <= GENESIS_EPOCH+1 {
		return
	}
	previousEpoch := state.PreviousEpoch()

	// epoch numbers are trusted, no errors
	previousBoundaryBlockRoot, _ := state.GetBlockRoot(previousEpoch)
	currentBoundaryBlockRoot, _ := state.GetBlockRoot(currentEpoch)

	// Get the sum balances of the boundary attesters
	previousEpochBoundaryAttestingBalance := Gwei(0)
	currentEpochBoundaryAttestingBalance := Gwei(0)
	for i, v := range state.Validators {
		vStatus := state.PrecomputedData.GetValidatorStatus(ValidatorIndex(i))
		if vStatus.Flags.HasMarkers(PrevEpochBoundaryAttester | UnslashedAttester) {
			previousEpochBoundaryAttestingBalance += v.EffectiveBalance
		}
		if vStatus.Flags.HasMarkers(CurrEpochBoundaryAttester | UnslashedAttester) {
			previousEpochBoundaryAttestingBalance += v.EffectiveBalance
		}
	}

	// Get the sum balances of the attesters for the epochs
	previousTotalBalance := state.Validators.GetTotalActiveEffectiveBalance(previousEpoch)
	currentTotalBalance := state.Validators.GetTotalActiveEffectiveBalance(currentEpoch)

	oldPreviousJustified := state.PreviousJustifiedCheckpoint
	oldCurrentJustified := state.CurrentJustifiedCheckpoint

	// Rotate current into previous
	state.PreviousJustifiedCheckpoint = state.CurrentJustifiedCheckpoint
	state.JustificationBits.NextEpoch()

	// > Justification
	if previousEpochBoundaryAttestingBalance*3 >= previousTotalBalance*2 {
		state.Justify(Checkpoint{
			Epoch: previousEpoch,
			Root:  previousBoundaryBlockRoot,
		})
	}
	if currentEpochBoundaryAttestingBalance*3 >= currentTotalBalance*2 {
		state.Justify(Checkpoint{
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
