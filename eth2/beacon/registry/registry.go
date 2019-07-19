package registry

import (
	. "github.com/protolambda/zrnt/eth2/beacon/validator"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
)

// Validator registry
type RegistryState struct {
	Validators ValidatorRegistry
	Balances   Balances
}

// Update effective balances with hysteresis
func (state *RegistryState) UpdateEffectiveBalances() {
	for i, v := range state.Validators {
		balance := state.Balances[i]
		if balance < v.EffectiveBalance ||
			v.EffectiveBalance+3*HALF_INCREMENT < balance {
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
	validator.ExitEpoch = state.Validators.ExitQueueEnd(currentEpoch)
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

type RegistryUpdateReq interface {
	VersioningMeta
	FinalityMeta
	ActivationExitMeta
}

func (state *RegistryState) ProcessEpochRegistryUpdates(meta RegistryUpdateReq) {
	// Process activation eligibility and ejections
	currentEpoch := meta.Epoch()
	for i, v := range state.Validators {
		if v.ActivationEligibilityEpoch == FAR_FUTURE_EPOCH &&
			v.EffectiveBalance == MAX_EFFECTIVE_BALANCE {
			v.ActivationEligibilityEpoch = currentEpoch
		}
		if v.IsActive(currentEpoch) &&
			v.EffectiveBalance <= EJECTION_BALANCE {
			state.InitiateValidatorExit(currentEpoch, ValidatorIndex(i))
		}
	}
	// Queue validators eligible for activation and not dequeued for activation prior to finalized epoch
	activationEpoch := meta.Finalized().Epoch.ComputeActivationExitEpoch()
	state.Validators.ProcessActivationQueue(activationEpoch, currentEpoch)
}
