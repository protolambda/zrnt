package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
)

func ProcessEpochSlashings(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	activeValidatorIndices := state.ValidatorRegistry.GetActiveValidatorIndices(currentEpoch)
	totalBalance := state.GetTotalBalanceOf(activeValidatorIndices)

	validatorCount := beacon.ValidatorIndex(len(state.ValidatorRegistry))
	for i := beacon.ValidatorIndex(0); i < validatorCount; i++ {
		v := &state.ValidatorRegistry[i]
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
			state.DecreaseBalance(i, penalty)
		}
	}
}
