package registry

import (
	. "github.com/protolambda/zrnt/eth2/beacon/validator"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

// Validator registry
type RegistryState struct {
	ValidatorsState
	BalancesState
}

// Update effective balances with hysteresis
func (state *RegistryState) UpdateEffectiveBalances() {
	const HYSTERESIS_INCREMENT = EFFECTIVE_BALANCE_INCREMENT / Gwei(HYSTERESIS_QUOTIENT)
	const DOWNWARD_THRESHOLD = HYSTERESIS_INCREMENT * Gwei(HYSTERESIS_DOWNWARD_MULTIPLIER)
	const UPWARD_THRESHOLD = HYSTERESIS_INCREMENT * Gwei(HYSTERESIS_UPWARD_MULTIPLIER)

	for i, v := range state.Validators {
		balance := state.Balances[i]
		if balance+DOWNWARD_THRESHOLD < v.EffectiveBalance ||
			v.EffectiveBalance+UPWARD_THRESHOLD < balance {
			v.EffectiveBalance = balance - (balance % EFFECTIVE_BALANCE_INCREMENT)
			if MAX_EFFECTIVE_BALANCE < v.EffectiveBalance {
				v.EffectiveBalance = MAX_EFFECTIVE_BALANCE
			}
		}
	}
}

// Initiate the exit of the validator of the given index
func (state *RegistryState) InitiateValidatorExit(currentEpoch Epoch, index ValidatorIndex) {
	validator := state.Validators[index]
	// Return if validator already initiated exit
	if validator.ExitEpoch != FAR_FUTURE_EPOCH {
		return
	}

	// Set validator exit epoch and withdrawable epoch
	validator.ExitEpoch = state.ExitQueueEnd(currentEpoch)
	validator.WithdrawableEpoch = validator.ExitEpoch + MIN_VALIDATOR_WITHDRAWABILITY_DELAY
}

func (state *RegistryState) AddNewValidator(pubkey BLSPubkey, withdrawalCreds Root, balance Gwei) {
	effBalance := balance - (balance % EFFECTIVE_BALANCE_INCREMENT)
	if effBalance > MAX_EFFECTIVE_BALANCE {
		effBalance = MAX_EFFECTIVE_BALANCE
	}
	validator := &Validator{
		Pubkey:                     pubkey,
		WithdrawalCredentials:      withdrawalCreds,
		ActivationEligibilityEpoch: FAR_FUTURE_EPOCH,
		ActivationEpoch:            FAR_FUTURE_EPOCH,
		ExitEpoch:                  FAR_FUTURE_EPOCH,
		WithdrawableEpoch:          FAR_FUTURE_EPOCH,
		EffectiveBalance:           effBalance,
	}
	state.Validators = append(state.Validators, validator)
	state.Balances = append(state.Balances, balance)
}

type RegistryUpdateEpochProcess interface {
	ProcessEpochRegistryUpdates()
}

type RegistryUpdatesFeature struct {
	State *RegistryState
	Meta  interface {
		meta.Versioning
		meta.Finality
		meta.ActivationExit
	}
}

func (f *RegistryUpdatesFeature) ProcessEpochRegistryUpdates() {
	// Process activation eligibility and ejections
	currentEpoch := f.Meta.CurrentEpoch()
	for i, v := range f.State.Validators {
		if v.IsEligibleForActivationQueue() {
			v.ActivationEligibilityEpoch = currentEpoch + 1
		}
		if v.IsActive(currentEpoch) &&
			v.EffectiveBalance <= EJECTION_BALANCE {
			f.State.InitiateValidatorExit(currentEpoch, ValidatorIndex(i))
		}
	}
	f.State.ProcessActivationQueue(currentEpoch, f.Meta.Finalized().Epoch)
}
