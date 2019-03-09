package slashing

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
)

func ProcessEpochSlashings(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	activeValidatorIndices := state.ValidatorRegistry.GetActiveValidatorIndices(currentEpoch)
	totalBalance := state.ValidatorBalances.GetTotalBalance(activeValidatorIndices)

	for index, validator := range state.ValidatorRegistry {
		if validator.Slashed &&
			currentEpoch == validator.WithdrawableEpoch-(beacon.LATEST_SLASHED_EXIT_LENGTH/2) {
			epochIndex := currentEpoch % beacon.LATEST_SLASHED_EXIT_LENGTH
			totalAtStart := state.LatestSlashedBalances[(epochIndex+1)%beacon.LATEST_SLASHED_EXIT_LENGTH]
			totalAtEnd := state.LatestSlashedBalances[epochIndex]
			balance := state.ValidatorBalances.GetEffectiveBalance(beacon.ValidatorIndex(index))
			state.ValidatorBalances[index] -= beacon.Max(
				balance*beacon.Min(
					(totalAtEnd-totalAtStart)*3,
					totalBalance,
				)/totalBalance,
				balance/beacon.MIN_PENALTY_QUOTIENT)
		}
	}
}
