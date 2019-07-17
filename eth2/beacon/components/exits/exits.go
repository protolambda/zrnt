package exits

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/meta"
	. "github.com/protolambda/zrnt/eth2/core"
)

type InitExitReq interface {
	VersioningMeta
	ValidatorMeta
	ActivationExitMeta
}

// Initiate the exit of the validator of the given index
func InitiateValidatorExit(meta InitExitReq, index ValidatorIndex) {
	validator := meta.Validator(index)
	// Return if validator already initiated exit
	if validator.ExitEpoch != FAR_FUTURE_EPOCH {
		return
	}

	// Set validator exit epoch and withdrawable epoch
	validator.ExitEpoch = meta.ExitQueueEnd(meta.Epoch())
	validator.WithdrawableEpoch = validator.ExitEpoch + MIN_VALIDATOR_WITHDRAWABILITY_DELAY
}
