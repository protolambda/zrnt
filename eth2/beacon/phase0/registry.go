package phase0

import (
	"context"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type RegistryIndices []common.ValidatorIndex

func (p *RegistryIndices) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*p)
		*p = append(*p, common.ValidatorIndex(0))
		return &((*p)[i])
	}, common.ValidatorIndexType.TypeByteLength(), spec.VALIDATOR_REGISTRY_LIMIT)
}

func (a RegistryIndices) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return a[i]
	}, common.ValidatorIndexType.TypeByteLength(), uint64(len(a)))
}

func (a RegistryIndices) ByteLength(spec *common.Spec) uint64 {
	return common.ValidatorIndexType.TypeByteLength() * uint64(len(a))
}

func (*RegistryIndices) FixedLength() uint64 {
	return 0
}

func (p RegistryIndices) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(p[i])
	}, uint64(len(p)), spec.VALIDATOR_REGISTRY_LIMIT)
}

type ValidatorRegistry []*Validator

func (a *ValidatorRegistry) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, &Validator{})
		return (*a)[i]
	}, ValidatorType.TypeByteLength(), spec.VALIDATOR_REGISTRY_LIMIT)
}

func (a ValidatorRegistry) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return a[i]
	}, ValidatorType.TypeByteLength(), uint64(len(a)))
}

func (a ValidatorRegistry) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(a)) * ValidatorType.TypeByteLength()
}

func (a *ValidatorRegistry) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li ValidatorRegistry) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return li[i]
		}
		return nil
	}, length, spec.VALIDATOR_REGISTRY_LIMIT)
}

func ValidatorsRegistryType(spec *common.Spec) ListTypeDef {
	return ComplexListType(ValidatorType, spec.VALIDATOR_REGISTRY_LIMIT)
}

type ValidatorsRegistryView struct{ *ComplexListView }

func AsValidatorsRegistry(v View, err error) (*ValidatorsRegistryView, error) {
	c, err := AsComplexList(v, err)
	return &ValidatorsRegistryView{c}, nil
}

func (registry *ValidatorsRegistryView) ValidatorCount() (uint64, error) {
	return registry.Length()
}

func (registry *ValidatorsRegistryView) Validator(index common.ValidatorIndex) (common.Validator, error) {
	return AsValidator(registry.Get(uint64(index)))
}

func (registry *ValidatorsRegistryView) Iter() (next func() (val common.Validator, ok bool, err error)) {
	iter := registry.ReadonlyIter()
	return func() (val common.Validator, ok bool, err error) {
		elem, ok, err := iter.Next()
		if err != nil || !ok {
			return nil, ok, err
		}
		v, err := AsValidator(elem, nil)
		return v, true, err
	}
}

func (registry *ValidatorsRegistryView) IsValidIndex(index common.ValidatorIndex) (valid bool, err error) {
	count, err := registry.Length()
	if err != nil {
		return false, err
	}
	return uint64(index) < count, nil
}

func ProcessEpochRegistryUpdates(ctx context.Context, spec *common.Spec, epc *EpochsContext, process *EpochProcess, state common.BeaconState) error {
	select {
	case <-ctx.Done():
		return common.TransitionCancelErr
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
			if err := val.SetWithdrawableEpoch(exitEnd + spec.MIN_VALIDATOR_WITHDRAWABILITY_DELAY); err != nil {
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
		dequeued := process.IndicesToMaybeActivate
		if uint64(len(dequeued)) > process.ChurnLimit {
			dequeued = dequeued[:process.ChurnLimit]
		}
		activationEpoch := spec.ComputeActivationExitEpoch(epc.CurrentEpoch.Epoch)
		for _, index := range dequeued {
			if process.Statuses[index].Validator.ActivationEligibilityEpoch > finality.Epoch {
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