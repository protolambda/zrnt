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
			penalty := Max(
				v.EffectiveBalance*Min(
					(totalAtEnd-totalAtStart)*3,
					totalBalance,
				)/totalBalance,
				v.EffectiveBalance/MIN_SLASHING_PENALTY_QUOTIENT)
			state.DecreaseBalance(ValidatorIndex(i), penalty)
		}
	}
}
