package epoch

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
)

func ProcessEpochSlashings(state *BeaconState) {
	currentEpoch := state.Epoch()
	totalBalance := state.Validators.GetTotalActiveEffectiveBalance(currentEpoch)

	epochIndex := currentEpoch % EPOCHS_PER_SLASHINGS_VECTOR
	// Compute slashed balances in the current epoch
	slashings := state.Slashings[(epochIndex+1)%EPOCHS_PER_SLASHINGS_VECTOR]

	for i, v := range state.Validators {
		if v.Slashed && currentEpoch+(EPOCHS_PER_SLASHINGS_VECTOR/2) == v.WithdrawableEpoch {
			scaledBalance := v.EffectiveBalance
			if balanceDiff := slashings * 3; totalBalance > balanceDiff {
				scaledBalance = (scaledBalance * balanceDiff) / totalBalance
			}
			penalty := scaledBalance
			if minimumPenalty := v.EffectiveBalance / MIN_SLASHING_PENALTY_QUOTIENT; minimumPenalty < penalty {
				penalty = minimumPenalty
			}
			state.Balances.DecreaseBalance(ValidatorIndex(i), penalty)
		}
	}
}
