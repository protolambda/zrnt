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

	previousEpochBoundaryAttesterIndices := state.Validators.FilterUnslashed(
		state.GetAttesters(state.PreviousEpochAttestations,
			func(att *AttestationData) bool {
				return att.Target.Root == previousBoundaryBlockRoot
			}))

	currentEpochBoundaryAttesterIndices := state.Validators.FilterUnslashed(
		state.GetAttesters(state.CurrentEpochAttestations,
			func(att *AttestationData) bool {
				return att.Target.Root == currentBoundaryBlockRoot
			}))

	oldPreviousJustified := state.PreviousJustifiedCheckpoint
	oldCurrentJustified := state.CurrentJustifiedCheckpoint

	// Rotate current into previous
	state.PreviousJustifiedCheckpoint = state.CurrentJustifiedCheckpoint
	state.PreviousJustifiedCheckpoint = state.CurrentJustifiedCheckpoint
	state.JustificationBits.NextEpoch()

	// Get the sum balances of the boundary attesters, and the total balance at the time.
	previousEpochBoundaryAttestingBalance := state.Validators.GetTotalEffectiveBalanceOf(previousEpochBoundaryAttesterIndices)
	previousTotalBalance := state.Validators.GetTotalEffectiveBalanceOf(state.Validators.GetActiveValidatorIndices(currentEpoch - 1))
	currentEpochBoundaryAttestingBalance := state.Validators.GetTotalEffectiveBalanceOf(currentEpochBoundaryAttesterIndices)
	currentTotalBalance := state.Validators.GetTotalEffectiveBalanceOf(state.Validators.GetActiveValidatorIndices(currentEpoch))

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
