package beacon

import (


)

// Validator registry
type RegistryState struct {
	//ValidatorsState
	//BalancesState
}

// Update effective balances with hysteresis
func (state *RegistryState) UpdateEffectiveBalances() error { // TODO
	//for i, v := range state.Validators {
	//	balance := state.Balances[i]
	//	if balance < v.EffectiveBalance ||
	//		v.EffectiveBalance+3*HALF_INCREMENT < balance {
	//		v.EffectiveBalance = balance - (balance % EFFECTIVE_BALANCE_INCREMENT)
	//		if MAX_EFFECTIVE_BALANCE < v.EffectiveBalance {
	//			v.EffectiveBalance = MAX_EFFECTIVE_BALANCE
	//		}
	//	}
	//}
	return nil
}

// Initiate the exit of the validator of the given index
func (state *RegistryState) InitiateValidatorExit(currentEpoch Epoch, index ValidatorIndex) error {
	//validator := state.Validators[index]
	//// Return if validator already initiated exit
	//if validator.ExitEpoch != FAR_FUTURE_EPOCH {
	//	return
	//}
	//
	//// Set validator exit epoch and withdrawable epoch
	//validator.ExitEpoch = state.ExitQueueEnd(currentEpoch)
	//validator.WithdrawableEpoch = validator.ExitEpoch + MIN_VALIDATOR_WITHDRAWABILITY_DELAY
	return nil
}

func (state *RegistryState) AddNewValidator(pubkey BLSPubkeyNode, withdrawalCreds Root, balance Gwei) error {
	//effBalance := balance - (balance % EFFECTIVE_BALANCE_INCREMENT)
	//if effBalance > MAX_EFFECTIVE_BALANCE {
	//	effBalance = MAX_EFFECTIVE_BALANCE
	//}
	//validator := &Validator{
	//	Pubkey:                     pubkey,
	//	WithdrawalCredentials:      withdrawalCreds,
	//	ActivationEligibilityEpoch: FAR_FUTURE_EPOCH,
	//	ActivationEpoch:            FAR_FUTURE_EPOCH,
	//	ExitEpoch:                  FAR_FUTURE_EPOCH,
	//	WithdrawableEpoch:          FAR_FUTURE_EPOCH,
	//	EffectiveBalance:           effBalance,
	//}
	//state.Validators = append(state.Validators, validator)
	//state.Balances = append(state.Balances, balance)
	return nil
}

type RegistryUpdateEpochProcess interface {
	ProcessEpochRegistryUpdates() error
}

type RegistryUpdatesFeature struct {
	State *RegistryState
	Meta  interface {
		meta.Versioning
		meta.Finality
		meta.ActivationExit
	}
}

func (f *RegistryUpdatesFeature) ProcessEpochRegistryUpdates() error {
	// Process activation eligibility and ejections
	//currentEpoch := f.Meta.CurrentEpoch()
	//for i, v := range f.State.Validators {
	//	if v.ActivationEligibilityEpoch == FAR_FUTURE_EPOCH &&
	//		v.EffectiveBalance == MAX_EFFECTIVE_BALANCE {
	//		v.ActivationEligibilityEpoch = currentEpoch
	//	}
	//	if v.IsActive(currentEpoch) &&
	//		v.EffectiveBalance <= EJECTION_BALANCE {
	//		f.State.InitiateValidatorExit(currentEpoch, ValidatorIndex(i))
	//	}
	//}
	//// Queue validators eligible for activation and not dequeued for activation prior to finalized epoch
	//activationEpoch := f.Meta.Finalized().Epoch.ComputeActivationExitEpoch()
	//f.State.ProcessActivationQueue(activationEpoch, currentEpoch)
	return nil
}
