package beacon

import (
	"context"
	"errors"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func (c *Phase0Config) BlockVoluntaryExits() ListTypeDef {
	return ListType(c.SignedVoluntaryExit(), c.MAX_VOLUNTARY_EXITS)
}

type VoluntaryExits struct {
	Items []SignedVoluntaryExit
	Limit uint64
}

func (li *VoluntaryExits) HashTreeRoot(hFn tree.HashFn) Root {
	length := uint64(len(li.Items))
	return hFn.Mixin(hFn.SeriesHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li.Items[i]
		}
		return nil
	}, length, li.Limit), length)
}

func (spec *Spec) ProcessVoluntaryExits(ctx context.Context, epc *EpochsContext, state *BeaconStateView, ops []SignedVoluntaryExit) error {
	for i := range ops {
		select {
		case <-ctx.Done():
			return TransitionCancelErr
		default: // Don't block.
			break
		}
		if err := spec.ProcessVoluntaryExit(state, epc, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

type VoluntaryExit struct {
	Epoch          Epoch // Earliest epoch when voluntary exit can be processed
	ValidatorIndex ValidatorIndex
}

func (v *VoluntaryExit) HashTreeRoot() Root {
	return ssz.HashTreeRoot(v, VoluntaryExitSSZ)
}

type SignedVoluntaryExit struct {
	Message   VoluntaryExit
	Signature BLSSignature
}

func (c *Phase0Config) VoluntaryExit() *ContainerTypeDef {
	return ContainerType("VoluntaryExit", []FieldDef{
		{"epoch", EpochType}, // Earliest epoch when voluntary exit can be processed
		{"validator_index", ValidatorIndexType},
	})
}

func (c *Phase0Config) SignedVoluntaryExit() *ContainerTypeDef {
	return ContainerType("SignedVoluntaryExit", []FieldDef{
		{"message", c.VoluntaryExit()},
		{"signature", BLSSignatureType},
	})
}

func (spec *Spec) ProcessVoluntaryExit(state *BeaconStateView, epc *EpochsContext, signedExit *SignedVoluntaryExit) error {
	exit := &signedExit.Message
	currentEpoch := epc.CurrentEpoch.Epoch
	if valid, err := state.IsValidIndex(exit.ValidatorIndex); err != nil {
		return err
	} else if !valid {
		return errors.New("invalid exit validator index")
	}
	vals, err := state.Validators()
	if err != nil {
		return err
	}
	validator, err := vals.Validator(exit.ValidatorIndex)
	if err != nil {
		return err
	}
	// Verify that the validator is active
	if isActive, err := spec.IsActive(validator, currentEpoch); err != nil {
		return err
	} else if !isActive {
		return errors.New("validator must be active to be able to voluntarily exit")
	}
	scheduledExitEpoch, err := validator.ExitEpoch()
	if err != nil {
		return err
	}
	// Verify exit has not been initiated
	if scheduledExitEpoch != FAR_FUTURE_EPOCH {
		return errors.New("validator already exited")
	}
	// Exits must specify an epoch when they become valid; they are not valid before then
	if currentEpoch < exit.Epoch {
		return errors.New("invalid exit epoch")
	}
	registeredActivationEpoch, err := validator.ActivationEpoch()
	if err != nil {
		return err
	}
	// Verify the validator has been active long enough
	if currentEpoch < registeredActivationEpoch+spec.SHARD_COMMITTEE_PERIOD {
		return errors.New("exit is too soon")
	}
	pubkey, ok := epc.PubkeyCache.Pubkey(exit.ValidatorIndex)
	if !ok {
		return errors.New("could not find index of exiting validator")
	}
	domain, err := state.GetDomain(spec.DOMAIN_VOLUNTARY_EXIT, exit.Epoch)
	if err != nil {
		return err
	}
	// Verify signature
	if !bls.Verify(
		pubkey,
		ComputeSigningRoot(signedExit.Message.HashTreeRoot(), domain),
		signedExit.Signature) {
		return errors.New("voluntary exit signature could not be verified")
	}
	// Initiate exit
	return spec.InitiateValidatorExit(epc, state, exit.ValidatorIndex)
}

// Initiate the exit of the validator of the given index
func (spec *Spec) InitiateValidatorExit(epc *EpochsContext, state *BeaconStateView, index ValidatorIndex) error {
	validators, err := state.Validators()
	if err != nil {
		return err
	}
	v, err := validators.Validator(index)
	if err != nil {
		return err
	}
	exitEp, err := v.ExitEpoch()
	if err != nil {
		return err
	}
	// Return if validator already initiated exit
	if exitEp != FAR_FUTURE_EPOCH {
		return nil
	}
	currentEpoch := epc.CurrentEpoch.Epoch

	// Set validator exit epoch and withdrawable epoch
	valIter := validators.ReadonlyIter()

	exitQueueEnd := spec.ComputeActivationExitEpoch(currentEpoch)
	exitQueueEndChurn := uint64(0)
	for {
		valContainer, ok, err := valIter.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		val, err := AsValidator(valContainer, nil)
		if err != nil {
			return err
		}
		valExit, err := val.ExitEpoch()
		if err != nil {
			return err
		}
		if valExit == FAR_FUTURE_EPOCH {
			continue
		}
		if valExit == exitQueueEnd {
			exitQueueEndChurn++
		} else if valExit > exitQueueEnd {
			exitQueueEnd = valExit
			exitQueueEndChurn = 1
		}
	}
	churnLimit := spec.GetChurnLimit(uint64(len(epc.CurrentEpoch.ActiveIndices)))
	if exitQueueEndChurn >= churnLimit {
		exitQueueEnd++
	}

	exitEp = exitQueueEnd
	if err := v.SetExitEpoch(exitEp); err != nil {
		return err
	}
	if err := v.SetWithdrawableEpoch(exitEp + spec.MIN_VALIDATOR_WITHDRAWABILITY_DELAY); err != nil {
		return err
	}
	return nil
}
