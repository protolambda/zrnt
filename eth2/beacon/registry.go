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

func (registry *ValidatorsRegistryView) ValidatorCount() (uint64, error) {
	return registry.Length()
}

func (registry *ValidatorsRegistryView) Validator(index ValidatorIndex) (*ValidatorView, error) {
	return AsValidator(registry.Get(uint64(index)))
}

func (state *BeaconStateView) ProcessEpochRegistryUpdates(epc *EpochsContext, process *EpochProcess) error {
	// Process activation eligibility and ejections
	currentEpoch := epc.CurrentEpoch.Epoch
	for i, v := range f.State.Validators {
		if v.ActivationEligibilityEpoch == FAR_FUTURE_EPOCH &&
			v.EffectiveBalance == MAX_EFFECTIVE_BALANCE {
			v.ActivationEligibilityEpoch = currentEpoch
		}
		if v.IsActive(currentEpoch) &&
			v.EffectiveBalance <= EJECTION_BALANCE {
			f.State.InitiateValidatorExit(currentEpoch, ValidatorIndex(i))
		}
	}
	// Queue validators eligible for activation and not dequeued for activation prior to finalized epoch
	activationEpoch := f.Meta.Finalized().Epoch.ComputeActivationExitEpoch()
	f.State.ProcessActivationQueue(activationEpoch, currentEpoch)


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
