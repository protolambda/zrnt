package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"sort"
)

func ProcessRegistryUpdates(state *BeaconState) {
	// Process activation eligibility and ejections
	currentEpoch := state.Epoch()
	for i, v := range state.ValidatorRegistry {
		balance := state.GetBalance(ValidatorIndex(i))
		if v.ActivationEligibilityEpoch == FAR_FUTURE_EPOCH && balance >= MAX_EFFECTIVE_BALANCE {
			v.ActivationEligibilityEpoch = currentEpoch
		}
		if v.IsActive(currentEpoch) && balance < EJECTION_BALANCE {
			state.InitiateValidatorExit(ValidatorIndex(i))
		}
	}
	// Queue validators eligible for activation and not dequeued for activation prior to finalized epoch
	activationQueue := make([]*Validator, 0)
	for i, v := range state.ValidatorRegistry {
		if v.ActivationEligibilityEpoch != FAR_FUTURE_EPOCH &&
			v.ActivationEpoch >= state.FinalizedEpoch.GetDelayedActivationExitEpoch() {
			activationQueue = append(activationQueue, v)
		}
	}
	sort.Slice(activationQueue, func(i int, j int) bool {
		return activationQueue[i].ActivationEligibilityEpoch <
			activationQueue[j].ActivationEligibilityEpoch
	})
	// Dequeued validators for activation up to churn limit (without resetting activation epoch)
	for _, v := range activationQueue[:state.GetChurnLimit()] {
		if v.ActivationEpoch == FAR_FUTURE_EPOCH {
			v.ActivationEpoch = currentEpoch.GetDelayedActivationExitEpoch()
		}
	}
}
