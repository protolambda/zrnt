package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
)

func ProcessEpochSlashings(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	activeValidatorIndices := state.ValidatorRegistry.GetActiveValidatorIndices(currentEpoch)
	totalBalance := state.Balances.GetTotalBalance(activeValidatorIndices)

	for index, validator := range state.ValidatorRegistry {
		if validator.Slashed &&
			currentEpoch == validator.WithdrawableEpoch-(beacon.LATEST_SLASHED_EXIT_LENGTH/2) {
			epochIndex := currentEpoch % beacon.LATEST_SLASHED_EXIT_LENGTH
			totalAtStart := state.LatestSlashedBalances[(epochIndex+1)%beacon.LATEST_SLASHED_EXIT_LENGTH]
			totalAtEnd := state.LatestSlashedBalances[epochIndex]
			balance := state.Balances.GetEffectiveBalance(beacon.ValidatorIndex(index))
			state.Balances[index] -= beacon.Max(
				balance*beacon.Min(
					(totalAtEnd-totalAtStart)*3,
					totalBalance,
				)/totalBalance,
				balance/beacon.MIN_PENALTY_QUOTIENT)
		}
	}
}
