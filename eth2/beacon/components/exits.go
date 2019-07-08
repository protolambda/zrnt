package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
)

func (state *BeaconState) GetChurnLimit() uint64 {
	return math.MaxU64(MIN_PER_EPOCH_CHURN_LIMIT,
		state.Validators.GetActiveValidatorCount(state.Epoch())/CHURN_LIMIT_QUOTIENT)
}

// Exit the validator with the given index
func (state *BeaconState) ExitValidator(index ValidatorIndex) {
	validator := state.Validators[index]
	// Update validator exit epoch if not previously exited
	if validator.ExitEpoch == FAR_FUTURE_EPOCH {
		validator.ExitEpoch = state.Epoch().GetDelayedActivationExitEpoch()
	}
}

// Initiate the exit of the validator of the given index
func (state *BeaconState) InitiateValidatorExit(index ValidatorIndex) {
	validator := state.Validators[index]
	// Return if validator already initiated exit
	if validator.ExitEpoch != FAR_FUTURE_EPOCH {
		return
	}
	// Compute exit queue epoch
	exitQueueEnd := state.Epoch().GetDelayedActivationExitEpoch()
	for _, v := range state.Validators {
		if v.ExitEpoch != FAR_FUTURE_EPOCH && v.ExitEpoch > exitQueueEnd {
			exitQueueEnd = v.ExitEpoch
		}
	}
	exitQueueChurn := uint64(0)
	for _, v := range state.Validators {
		if v.ExitEpoch == exitQueueEnd {
			exitQueueChurn++
		}
	}
	if exitQueueChurn >= state.GetChurnLimit() {
		exitQueueEnd++
	}

	// Set validator exit epoch and withdrawable epoch
	validator.ExitEpoch = exitQueueEnd
	validator.WithdrawableEpoch = validator.ExitEpoch + MIN_VALIDATOR_WITHDRAWABILITY_DELAY
}
