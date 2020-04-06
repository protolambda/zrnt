package beacon

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

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
	currentEpoch, err := input.CurrentEpoch()
	if err != nil {
		return err
	}
	if valid, err := input.IsValidIndex(exit.ValidatorIndex); err != nil {
		return err
	} else if !valid {
		return errors.New("invalid exit validator index")
	}
	// Verify that the validator is active
	if isActive, err := input.IsActive(exit.ValidatorIndex, currentEpoch); err != nil {
		return err
	} else if !isActive {
		return errors.New("validator must be active to be able to voluntarily exit")
	}
	scheduledExitEpoch, err := input.ExitEpoch(exit.ValidatorIndex)
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
	registeredActivationEpoch, err := input.ActivationEpoch(exit.ValidatorIndex)
	if err != nil {
		return err
	}
	// Verify the validator has been active long enough
	if currentEpoch < registeredActivationEpoch+PERSISTENT_COMMITTEE_PERIOD {
		return errors.New("exit is too soon")
	}
	pubkey, err := input.Pubkey(exit.ValidatorIndex)
	if err != nil {
		return err
	}
	domain, err := input.GetDomain(DOMAIN_VOLUNTARY_EXIT, exit.Epoch)
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
	return state.InitiateValidatorExit(epc, currentEpoch, exit.ValidatorIndex)
}

// Initiate the exit of the validator of the given index
func (state *BeaconStateView) InitiateValidatorExit(epc *EpochsContext, currentEpoch Epoch, index ValidatorIndex) error {
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
