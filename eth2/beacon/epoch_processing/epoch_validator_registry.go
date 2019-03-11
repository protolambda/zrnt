package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/math"
)

func ProcessEpochValidatorRegistry(state *beacon.BeaconState) {
	state.PreviousShufflingEpoch = state.CurrentShufflingEpoch
	state.PreviousShufflingStartShard = state.CurrentShufflingStartShard
	state.PreviousShufflingSeed = state.CurrentShufflingSeed
	currentEpoch := state.Epoch()
	nextEpoch := currentEpoch + 1
	// Check if we should update, and if so, update
	if shouldUpdateValidatorRegistry(state) {
		// update!
		UpdateValidatorRegistry(state)
		// If we update the registry, update the shuffling data and shards as well
		state.CurrentShufflingEpoch = nextEpoch
		commiteeCount := state.GetCurrentEpochCommitteeCount()
		shard := (state.CurrentShufflingStartShard + beacon.Shard(commiteeCount)) % beacon.SHARD_COUNT
		state.CurrentShufflingStartShard = shard
		state.CurrentShufflingSeed = state.GenerateSeed(state.CurrentShufflingEpoch)
	} else {
		// If processing at least one crosslink keeps failing, then reshuffle every power of two,
		// but don't update the current_shuffling_start_shard
		epochsSinceLastRegistryUpdate := currentEpoch - state.ValidatorRegistryUpdateEpoch
		if epochsSinceLastRegistryUpdate > 1 && math.IsPowerOfTwo(uint64(epochsSinceLastRegistryUpdate)) {
			state.CurrentShufflingEpoch = nextEpoch
			state.CurrentShufflingSeed = state.GenerateSeed(state.CurrentShufflingEpoch)
		}
	}
}

func shouldUpdateValidatorRegistry(state *beacon.BeaconState) bool {
	// Must have finalized a new block
	if state.FinalizedEpoch <= state.ValidatorRegistryUpdateEpoch {
		return false
	}
	// Must have processed new crosslinks on all shards of the current epoch
	commiteeCount := state.GetCurrentEpochCommitteeCount()
	for i := uint64(0); i < commiteeCount; i++ {
		shard := (state.CurrentShufflingStartShard + beacon.Shard(i)) % beacon.SHARD_COUNT
		if state.LatestCrosslinks[shard].Epoch <= state.ValidatorRegistryUpdateEpoch {
			return false
		}
	}
	return true
}

// Update validator registry.
func UpdateValidatorRegistry(state *beacon.BeaconState) {
	currentEpoch := state.Epoch()
	activeValidatorIndices := state.ValidatorRegistry.GetActiveValidatorIndices(currentEpoch)
	totalBalance := state.ValidatorBalances.GetTotalBalance(activeValidatorIndices)
	maxBalanceChurn := beacon.Max(
		beacon.MAX_DEPOSIT_AMOUNT,
		totalBalance/(2*beacon.MAX_BALANCE_CHURN_QUOTIENT))
	// Activate validators within the allowable balance churn
	balanceChurn := beacon.Gwei(0)
	for i := 0; i < len(state.ValidatorRegistry); i++ {
		v := &state.ValidatorRegistry[i]
		if v.ActivationEpoch == beacon.FAR_FUTURE_EPOCH &&
			state.ValidatorBalances[i] >= beacon.MAX_DEPOSIT_AMOUNT {
			// Check the balance churn would be within the allowance
			balanceChurn += state.ValidatorBalances.GetEffectiveBalance(beacon.ValidatorIndex(i))
			if balanceChurn > maxBalanceChurn {
				break
			}

			// Activate validator
			state.ActivateValidator(beacon.ValidatorIndex(i), false)
		}
	}
	// Exit validators within the allowable balance churn
	balanceChurn = 0
	for i := 0; i < len(state.ValidatorRegistry); i++ {
		v := &state.ValidatorRegistry[i]
		if v.ExitEpoch == beacon.FAR_FUTURE_EPOCH && v.InitiatedExit {
			// Check the balance churn would be within the allowance
			balanceChurn += state.ValidatorBalances.GetEffectiveBalance(beacon.ValidatorIndex(i))
			if balanceChurn > maxBalanceChurn {
				break
			}

			// Exit validator
			state.ExitValidator(beacon.ValidatorIndex(i))
		}
	}
	state.ValidatorRegistryUpdateEpoch = currentEpoch
}
