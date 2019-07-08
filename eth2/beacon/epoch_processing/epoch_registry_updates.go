package epoch_processing

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
	"sort"
)

func ProcessEpochRegistryUpdates(state *BeaconState) {
	// Process activation eligibility and ejections
	currentEpoch := state.Epoch()
	for i, v := range state.Validators {
		if v.ActivationEligibilityEpoch == FAR_FUTURE_EPOCH &&
			v.EffectiveBalance >= MAX_EFFECTIVE_BALANCE {
			v.ActivationEligibilityEpoch = currentEpoch
		}
		if v.IsActive(currentEpoch) &&
			v.EffectiveBalance <= EJECTION_BALANCE {
			state.InitiateValidatorExit(ValidatorIndex(i))
		}
	}
	// Queue validators eligible for activation and not dequeued for activation prior to finalized epoch
	activationQueue := make([]*Validator, 0)
	for _, v := range state.Validators {
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
	queueLen := uint64(len(activationQueue))
	if churnLimit := state.GetChurnLimit(); churnLimit < queueLen {
		queueLen = churnLimit
	}
	for _, v := range activationQueue[:queueLen] {
		if v.ActivationEpoch == FAR_FUTURE_EPOCH {
			v.ActivationEpoch = currentEpoch.GetDelayedActivationExitEpoch()
		}
	}
}
