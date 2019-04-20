package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
)

func ProcessEpochSlashings(state *BeaconState) {
	currentEpoch := state.Epoch()
	activeValidatorIndices := state.ValidatorRegistry.GetActiveValidatorIndices(currentEpoch)
	totalBalance := state.GetTotalBalanceOf(activeValidatorIndices)

	for i, v := range state.ValidatorRegistry {
		if v.Slashed &&
			currentEpoch == v.WithdrawableEpoch-(LATEST_SLASHED_EXIT_LENGTH/2) {
			epochIndex := currentEpoch % LATEST_SLASHED_EXIT_LENGTH
			totalAtStart := state.LatestSlashedBalances[(epochIndex+1)%LATEST_SLASHED_EXIT_LENGTH]
			totalAtEnd := state.LatestSlashedBalances[epochIndex]
			balance := state.GetEffectiveBalance(ValidatorIndex(i))
			penalty := Max(
				balance*Min(
					(totalAtEnd-totalAtStart)*3,
					totalBalance,
				)/totalBalance,
				balance/MIN_PENALTY_QUOTIENT)
			state.DecreaseBalance(ValidatorIndex(i), penalty)
		}
	}
}
