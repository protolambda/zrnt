package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
)

func ProcessEpochJustification(state *beacon.BeaconState) {

	currentEpoch := state.Epoch()
	// epoch numbers are trusted, no errors
	previousBoundaryBlockRoot, _ := state.GetBlockRoot((currentEpoch - 1).GetStartSlot())
	currentBoundaryBlockRoot, _ := state.GetBlockRoot(currentEpoch.GetStartSlot())

	previousEpochBoundaryAttesterIndices := make([]beacon.ValidatorIndex, 0)
	currentEpochBoundaryAttesterIndices := make([]beacon.ValidatorIndex, 0)
	for _, att := range state.PreviousEpochAttestations {
		// If the attestation is for the boundary:
		if att.Data.TargetRoot == previousBoundaryBlockRoot {
			participants, _ := state.GetAttestationParticipants(&att.Data, &att.AggregationBitfield)
			for _, vIndex := range participants {
				previousEpochBoundaryAttesterIndices = append(previousEpochBoundaryAttesterIndices, vIndex)
			}
		}
	}
	for _, att := range state.CurrentEpochAttestations {
		// If the attestation is for the boundary:
		if att.Data.TargetRoot == currentBoundaryBlockRoot {
			participants, _ := state.GetAttestationParticipants(&att.Data, &att.AggregationBitfield)
			for _, vIndex := range participants {
				currentEpochBoundaryAttesterIndices = append(currentEpochBoundaryAttesterIndices, vIndex)
			}
		}
	}

	newJustifiedEpoch := state.CurrentJustifiedEpoch
	newFinalizedEpoch := state.FinalizedEpoch
	// Rotate the justification bitfield up one epoch to make room for the current epoch
	state.JustificationBitfield <<= 1

	// Get the sum balances of the boundary attesters, and the total balance at the time.
	previousEpochBoundaryAttestingBalance := state.ValidatorBalances.GetTotalBalance(previousEpochBoundaryAttesterIndices)
	previousTotalBalance := state.ValidatorBalances.GetTotalBalance(state.ValidatorRegistry.GetActiveValidatorIndices(currentEpoch - 1))
	currentEpochBoundaryAttestingBalance := state.ValidatorBalances.GetTotalBalance(currentEpochBoundaryAttesterIndices)
	currentTotalBalance := state.ValidatorBalances.GetTotalBalance(state.ValidatorRegistry.GetActiveValidatorIndices(currentEpoch))

	// > Justification
	// If the previous epoch gets justified, fill the second last bit
	if 3*previousEpochBoundaryAttestingBalance >= 2*previousTotalBalance {
		state.JustificationBitfield |= 2
		newJustifiedEpoch = currentEpoch - 1
	}
	// If the current epoch gets justified, fill the last bit
	if 3*currentEpochBoundaryAttestingBalance >= 2*currentTotalBalance {
		state.JustificationBitfield |= 1
		newJustifiedEpoch = currentEpoch
	}
	// > Finalization
	// The 2nd/3rd/4th most recent epochs are all justified, the 2nd using the 4th as source
	if (state.JustificationBitfield>>1)&7 == 7 && state.PreviousJustifiedEpoch == currentEpoch-3 {
		newFinalizedEpoch = state.PreviousJustifiedEpoch
	}
	// The 2nd/3rd most recent epochs are both justified, the 2nd using the 3rd as source
	if (state.JustificationBitfield>>1)&3 == 3 && state.PreviousJustifiedEpoch == currentEpoch-2 {
		newFinalizedEpoch = state.PreviousJustifiedEpoch
	}
	// The 1st/2nd/3rd most recent epochs are all justified, the 1st using the 3rd as source
	if (state.JustificationBitfield>>0)&7 == 7 && state.CurrentJustifiedEpoch == currentEpoch-2 {
		newFinalizedEpoch = state.CurrentJustifiedEpoch
	}
	// The 1st/2nd most recent epochs are both justified, the 1st using the 2nd as source
	if (state.JustificationBitfield>>0)&3 == 3 && state.CurrentJustifiedEpoch == currentEpoch-1 {
		newFinalizedEpoch = state.CurrentJustifiedEpoch
	}
	// Rotate justified epochs
	state.PreviousJustifiedEpoch = state.CurrentJustifiedEpoch
	state.PreviousJustifiedRoot = state.CurrentJustifiedRoot
	// Update current state justification/finality fields
	if newJustifiedEpoch != state.CurrentJustifiedEpoch {
		state.CurrentJustifiedEpoch = newJustifiedEpoch
		root, err := state.GetBlockRoot(newJustifiedEpoch.GetStartSlot())
		if err != nil {
			panic(err)
		}
		state.CurrentJustifiedRoot = root
	}
	if newFinalizedEpoch != state.FinalizedEpoch {
		state.FinalizedEpoch = newFinalizedEpoch
		root, err := state.GetBlockRoot(newFinalizedEpoch.GetStartSlot())
		if err != nil {
			panic(err)
		}
		state.FinalizedRoot = root
	}

	state.CurrentJustifiedEpoch = newJustifiedEpoch
}
