package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
)

func ProcessEpochSlashings(state *BeaconState) {
	currentEpoch := state.Epoch()
	totalBalance := state.GetTotalActiveBalance()

	epochIndex := currentEpoch % LATEST_SLASHED_EXIT_LENGTH
	// Compute slashed balances in the current epoch
	totalAtStart := state.LatestSlashedBalances[(epochIndex+1)%LATEST_SLASHED_EXIT_LENGTH]
	totalAtEnd := state.LatestSlashedBalances[epochIndex]

	for i, v := range state.ValidatorRegistry {
		if v.Slashed && currentEpoch == v.WithdrawableEpoch-(LATEST_SLASHED_EXIT_LENGTH/2) {
			scaledBalance := v.EffectiveBalance
			if balanceDiff := (totalAtEnd-totalAtStart)*3; totalBalance > balanceDiff {
				scaledBalance = (scaledBalance * balanceDiff) / totalBalance
			}
			penalty := scaledBalance
			if minimumPenalty := v.EffectiveBalance/MIN_SLASHING_PENALTY_QUOTIENT; minimumPenalty < penalty {
				penalty = minimumPenalty
			}
			state.DecreaseBalance(ValidatorIndex(i), penalty)
		}
	}
}
