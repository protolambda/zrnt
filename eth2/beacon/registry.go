package beacon

import (
	"context"
	. "github.com/protolambda/ztyp/view"
)

type ValidatorRegistry []*Validator

func (_ *ValidatorRegistry) Limit() uint64 {
	return VALIDATOR_REGISTRY_LIMIT
}

var ValidatorsRegistryType = ComplexListType(ValidatorType, VALIDATOR_REGISTRY_LIMIT)

type ValidatorsRegistryView struct{ *ComplexListView }

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

func (state *BeaconStateView) ProcessEpochRegistryUpdates(ctx context.Context, epc *EpochsContext, process *EpochProcess) error {
	select {
	case <-ctx.Done():
		return TransitionCancelErr
	default: // Don't block.
		break
	}
	vals, err := state.Validators()
	if err != nil {
		return err
	}
	// process ejections
	{
		exitEnd := process.ExitQueueEnd
		endChurn := process.ExitQueueEndChurn
		for _, index := range process.IndicesToEject {
			val, err := vals.Validator(index)
			if err != nil {
				return err
			}
			if err := val.SetExitEpoch(exitEnd); err != nil {
				return err
			}
			if err := val.SetWithdrawableEpoch(exitEnd + MIN_VALIDATOR_WITHDRAWABILITY_DELAY); err != nil {
				return err
			}
			endChurn += 1
			if endChurn >= process.ChurnLimit {
				endChurn = 0
				exitEnd += 1
			}
		}
	}

	// Process activation eligibility
	{
		eligibilityEpoch := epc.CurrentEpoch.Epoch + 1
		for _, index := range process.IndicesToSetActivationEligibility {
			val, err := vals.Validator(index)
			if err != nil {
				return err
			}
			if err := val.SetActivationEligibilityEpoch(eligibilityEpoch); err != nil {
				return err
			}
		}
	}
	// Process activations
	{
		finality, err := state.FinalizedCheckpoint()
		if err != nil {
			return err
		}
		finalizedEpoch, err := finality.Epoch()
		if err != nil {
			return err
		}
		dequeued := process.IndicesToMaybeActivate
		if uint64(len(dequeued)) > process.ChurnLimit {
			dequeued = dequeued[:process.ChurnLimit]
		}
		activationEpoch := epc.CurrentEpoch.Epoch.ComputeActivationExitEpoch()
		for _, index := range dequeued {
			if process.Statuses[index].Validator.ActivationEligibilityEpoch > finalizedEpoch {
				// remaining validators all have an activation_eligibility_epoch that is higher anyway, break early
				// The tie-breaks were already sorted correctly in the IndicesToMaybeActivate queue.
				break
			}
			val, err := vals.Validator(index)
			if err != nil {
				return err
			}
			if err := val.SetActivationEpoch(activationEpoch); err != nil {
				return err
			}
		}
	}
	return nil
}
