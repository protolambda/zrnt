package registry

import (
	. "github.com/protolambda/zrnt/eth2/beacon/validator"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
)

var ValidatorsRegistryType = ListType(ValidatorType, VALIDATOR_REGISTRY_LIMIT)

type ValidatorsRegistry struct { *ListView }

func (state *ValidatorsRegistry) IsValidIndex(index ValidatorIndex) (bool, error) {
	count, err := state.ValidatorCount()
	return index < ValidatorIndex(count), err
}

func (state *ValidatorsRegistry) ValidatorCount() (uint64, error) {
	return state.Length()
}

func (state *ValidatorsRegistry) Validator(index ValidatorIndex) (*Validator, error) {
	v, err := ContainerProp(PropReader(state, uint64(index))).Container()
	if err != nil {
		return nil, err
	}
	return &Validator{ContainerView: v}, nil
}

func (state *ValidatorsRegistry) Pubkey(index ValidatorIndex) (BLSPubkey, error) {
	v, err := state.Validator(index)
	if err != nil {
		return BLSPubkey{}, err
	}
	return v.Pubkey()
}

// TODO: probably really slow, should have a pubkey cache or something
func (state *ValidatorsRegistry) ValidatorIndex(pubkey BLSPubkey) (index ValidatorIndex, exists bool, err error) {
	count, err := state.ValidatorCount()
	if err != nil {
		return 0, false, err
	}
	for i := ValidatorIndex(0); i < ValidatorIndex(count); i++ {
		v, err := state.Validator(i)
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

func (state *ValidatorsRegistry) WithdrawableEpoch(index ValidatorIndex) (Epoch, error) {
	v, err := state.Validator(index)
	if err != nil {
		return 0, err
	}
	return v.WithdrawableEpoch()
}

func (state *ValidatorsRegistry) IsActive(index ValidatorIndex, epoch Epoch) (bool, error) {
	v, err := state.Validator(index)
	if err != nil {
		return false, err
	}
	return v.IsActive(epoch)
}

func (state *ValidatorsRegistry) SlashAndDelayWithdraw(index ValidatorIndex, withdrawalEpoch Epoch) error {
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

func (state *ValidatorsRegistry) GetActiveValidatorIndices(epoch Epoch) (RegistryIndices, error) {
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

func (state *ValidatorsRegistry) ComputeActiveIndexRoot(epoch Epoch) (Root, error) {
	indices, err := state.GetActiveValidatorIndices(epoch)
	if err != nil {
		return Root{}, err
	}
	return indices.HashTreeRoot(), nil
}

func (state *ValidatorsRegistry) GetActiveValidatorCount(epoch Epoch) (activeCount uint64, err error) {
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

func (state *ValidatorsRegistry) IsSlashed(index ValidatorIndex) (bool, error) {
	v, err := state.Validator(index)
	if err != nil {
		return false, err
	}
	return v.Slashed()
}

func (state *ValidatorsRegistry) ProcessActivationQueue(currentEpoch Epoch) error {
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

func (state *ValidatorsRegistry) EffectiveBalance(index ValidatorIndex) (Gwei, error) {
	v, err := state.Validator(index)
	if err != nil {
		return 0, err
	}
	return v.EffectiveBalance()
}
