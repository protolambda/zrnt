package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
)

func ProcessEpochSlashings(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	activeValidatorIndices := state.ValidatorRegistry.GetActiveValidatorIndices(currentEpoch)
	totalBalance := state.GetTotalBalanceOf(activeValidatorIndices)

	for i, v := range state.ValidatorRegistry {
		if v.Slashed &&
			currentEpoch == v.WithdrawableEpoch-(beacon.LATEST_SLASHED_EXIT_LENGTH/2) {
			epochIndex := currentEpoch % beacon.LATEST_SLASHED_EXIT_LENGTH
			totalAtStart := state.LatestSlashedBalances[(epochIndex+1)%beacon.LATEST_SLASHED_EXIT_LENGTH]
			totalAtEnd := state.LatestSlashedBalances[epochIndex]
			balance := state.GetEffectiveBalance(beacon.ValidatorIndex(i))
			penalty := beacon.Max(
				balance*beacon.Min(
					(totalAtEnd-totalAtStart)*3,
					totalBalance,
				)/totalBalance,
				balance/beacon.MIN_PENALTY_QUOTIENT)
			state.DecreaseBalance(beacon.ValidatorIndex(i), penalty)
		}
	}
}
