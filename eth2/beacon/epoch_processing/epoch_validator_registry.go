package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
)

func ProcessEpochValidatorRegistry(state *beacon.BeaconState) {
	// Check if we should update, and if so, update
	if state.FinalizedEpoch > state.ValidatorRegistryUpdateEpoch {
		UpdateValidatorRegistry(state)
	}
	state.LatestStartShard = (
		state.LatestStartShard +
		beacon.Shard(state.GetCurrentEpochCommitteeCount())) % beacon.SHARD_COUNT
}

// Update validator registry.
func UpdateValidatorRegistry(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	activeValidatorIndices := state.ValidatorRegistry.GetActiveValidatorIndices(currentEpoch)
	totalBalance := state.GetTotalBalanceOf(activeValidatorIndices)
	maxBalanceChurn := beacon.Max(
		beacon.MAX_DEPOSIT_AMOUNT,
		totalBalance/(2*beacon.MAX_BALANCE_CHURN_QUOTIENT))
	// Activate validators within the allowable balance churn
	balanceChurn := beacon.Gwei(0)
	for i, v := range state.ValidatorRegistry {
		if v.ActivationEpoch == beacon.FAR_FUTURE_EPOCH &&
			state.GetBalance(beacon.ValidatorIndex(i)) >= beacon.MAX_DEPOSIT_AMOUNT {
			// Check the balance churn would be within the allowance
			balanceChurn += state.GetEffectiveBalance(beacon.ValidatorIndex(i))
			if balanceChurn > maxBalanceChurn {
				break
			}

			// Activate validator
			state.ActivateValidator(beacon.ValidatorIndex(i), false)
		}
	}
	// Exit validators within the allowable balance churn
	balanceChurn = state.LatestSlashedBalances[state.ValidatorRegistryUpdateEpoch % beacon.LATEST_SLASHED_EXIT_LENGTH] -
			state.LatestSlashedBalances[currentEpoch % beacon.LATEST_SLASHED_EXIT_LENGTH]
	for i, v := range state.ValidatorRegistry {
		if v.ExitEpoch == beacon.FAR_FUTURE_EPOCH && v.InitiatedExit {
			// Check the balance churn would be within the allowance
			balanceChurn += state.GetEffectiveBalance(beacon.ValidatorIndex(i))
			if balanceChurn > maxBalanceChurn {
				break
			}

			// Exit validator
			state.ExitValidator(beacon.ValidatorIndex(i))
		}
	}
	state.ValidatorRegistryUpdateEpoch = currentEpoch
}
