package beacon

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

var VoluntaryExitsType = ListType(SignedVoluntaryExitType, MAX_VOLUNTARY_EXITS)

type VoluntaryExits []SignedVoluntaryExit

func (*VoluntaryExits) Limit() uint64 {
	return MAX_VOLUNTARY_EXITS
}

func (state *BeaconStateView) ProcessVoluntaryExits(epc *EpochsContext, ops []SignedVoluntaryExit) error {
	for i := range ops {
		if err := state.ProcessVoluntaryExit(epc, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

var VoluntaryExitSSZ = zssz.GetSSZ((*VoluntaryExit)(nil))

type VoluntaryExit struct {
	Epoch          Epoch // Earliest epoch when voluntary exit can be processed
	ValidatorIndex ValidatorIndex
}

func (v *VoluntaryExit) HashTreeRoot() Root {
	return ssz.HashTreeRoot(v, VoluntaryExitSSZ)
}

var SignedVoluntaryExitSSZ = zssz.GetSSZ((*SignedVoluntaryExit)(nil))

type SignedVoluntaryExit struct {
	Message VoluntaryExit
	Signature BLSSignature
}

var VoluntaryExitType = ContainerType("VoluntaryExit", []FieldDef{
	{"epoch", EpochType}, // Earliest epoch when voluntary exit can be processed
	{"validator_index", ValidatorIndexType},
})

var SignedVoluntaryExitType = ContainerType("SignedVoluntaryExit", []FieldDef{
	{"message", VoluntaryExitType},
	{"signature", BLSSignatureType},
})

func (state *BeaconStateView) ProcessVoluntaryExit(epc *EpochsContext, signedExit *SignedVoluntaryExit) error {
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
	if isActive, err := validator.IsActive(currentEpoch); err != nil {
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
	if currentEpoch < registeredActivationEpoch+PERSISTENT_COMMITTEE_PERIOD {
		return errors.New("exit is too soon")
	}
	pubkey, ok := epc.PubkeyCache.Pubkey(exit.ValidatorIndex)
	if !ok {
		return errors.New("could not find index of exiting validator")
	}
	domain, err := state.GetDomain(DOMAIN_VOLUNTARY_EXIT, exit.Epoch)
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
	return state.InitiateValidatorExit(epc, exit.ValidatorIndex)
}

// Initiate the exit of the validator of the given index
func (state *BeaconStateView) InitiateValidatorExit(epc *EpochsContext, index ValidatorIndex) error {
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

	exitQueueEnd := currentEpoch.ComputeActivationExitEpoch()
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
	churnLimit := GetChurnLimit(uint64(len(epc.CurrentEpoch.ActiveIndices)))
	if exitQueueEndChurn >= churnLimit {
		exitQueueEnd++
	}

	exitEp = exitQueueEnd
	if err := v.SetExitEpoch(exitEp); err != nil {
		return err
	}
	if err := v.SetWithdrawableEpoch(exitEp + MIN_VALIDATOR_WITHDRAWABILITY_DELAY); err != nil {
		return err
	}
	return nil
}
