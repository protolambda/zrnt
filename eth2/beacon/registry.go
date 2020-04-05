package beacon

import (

	. "github.com/protolambda/ztyp/view"
)

var ValidatorsRegistryType = ComplexListType(ValidatorType, VALIDATOR_REGISTRY_LIMIT)

type ValidatorsRegistryView struct { *ComplexListView }

func AsValidatorsRegistry(v View, err error) (*ValidatorsRegistryView, error) {
	c, err := AsComplexList(v, err)
	return &ValidatorsRegistryView{c}, nil
}

func (registry *ValidatorsRegistryView) IsValidIndex(index ValidatorIndex) (bool, error) {
	count, err := registry.ValidatorCount()
	return index < ValidatorIndex(count), err
}

func (registry *ValidatorsRegistryView) ValidatorCount() (uint64, error) {
	return registry.Length()
}

func (registry *ValidatorsRegistryView) Validator(index ValidatorIndex) (*ValidatorView, error) {
	return AsValidator(registry.Get(uint64(index)))
}

// TODO: probably really slow, should have a pubkey cache or something
func (registry *ValidatorsRegistryView) ValidatorIndex(pubkey BLSPubkey) (index ValidatorIndex, exists bool, err error) {
	count, err := registry.ValidatorCount()
	if err != nil {
		return 0, false, err
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := registry.Validator(i)
		if err != nil {
			return 0, false, err
		}
		valPub, err := v.Pubkey()
		if err != nil {
			return 0, false, err
		}
		if valPub == pubkey {
			return i, true, nil
		}
	}
	return 0, false, err
}

func (state *ValidatorsRegistryView) SlashAndDelayWithdraw(index ValidatorIndex, withdrawalEpoch Epoch) error {
	v, err := state.Validator(index)
	if err != nil {
		return err
	}
	if err := v.MakeSlashed(); err != nil {
		return err
	}
	prevWithdrawalEpoch, err := v.WithdrawableEpoch()
	if err != nil {
		return err
	}
	if withdrawalEpoch > prevWithdrawalEpoch {
		if err := v.SetWithdrawableEpoch(withdrawalEpoch); err != nil {
			return err
		}
	}
	return nil
}

func (state *ValidatorsRegistryView) GetActiveValidatorIndices(epoch Epoch) (RegistryIndices, error) {
	count, err := state.ValidatorCount()
	if err != nil {
		return nil, err
	}
	res := make(RegistryIndices, 0, count)
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return nil, err
		}
		if active, err := v.IsActive(epoch); err != nil {
			return nil, err
		} else if active {
			res = append(res, i)
		}
	}
	return res, nil
}

func (state *ValidatorsRegistryView) ComputeActiveIndexRoot(epoch Epoch) (Root, error) {
	indices, err := state.GetActiveValidatorIndices(epoch)
	if err != nil {
		return Root{}, err
	}
	return indices.HashTreeRoot(), nil
}

func (state *ValidatorsRegistryView) GetActiveValidatorCount(epoch Epoch) (activeCount uint64, err error) {
	count, err := state.ValidatorCount()
	if err != nil {
		return 0, err
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
		if err != nil {
			return 0, err
		}
		if active, err := v.IsActive(epoch); err != nil {
			return 0, err
		} else if active {
			activeCount++
		}
	}
	return
}

func (state *ValidatorsRegistryView) ProcessActivationQueue(currentEpoch Epoch) error {
	// Dequeued validators for activation up to churn limit (without resetting activation epoch)
	queueLen := uint64(len(activationQueue))
	if churnLimit, err := state.GetChurnLimit(); err != nil {
		return err
	} else if churnLimit < queueLen {
		queueLen = churnLimit
	}

	for _, item := range activationQueue[:queueLen] {
		if item.activation == FAR_FUTURE_EPOCH {
			v, err := state.Validator(item.valIndex)
			if err != nil {
				return err
			}
			if err := v.SetActivationEpoch(currentEpoch.ComputeActivationExitEpoch()); err != nil {
				return err
			}
		}
	}
	return nil
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

func (state *BeaconStateView) ProcessEpochRegistryUpdates() error {
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
