package exits

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

type VoluntaryExitProcessor interface {
	ProcessVoluntaryExits(input VoluntaryExitProcessInput, ops []SignedVoluntaryExit) error
	ProcessVoluntaryExit(input VoluntaryExitProcessInput, signedExit *SignedVoluntaryExit) error
}

type VoluntaryExitProcessInput interface {
	meta.ActiveIndices
	meta.Pubkeys
	meta.SigDomain
	meta.ExitEpoch
	meta.ActivationEpoch
	meta.Versioning
	meta.RegistrySize
	meta.Exits
}

func ProcessVoluntaryExits(input VoluntaryExitProcessInput, ops []SignedVoluntaryExit) error {
	for i := range ops {
		if err := ProcessVoluntaryExit(input, &ops[i]); err != nil {
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

var VoluntaryExitType = &ContainerType{
	{"epoch", EpochType}, // Earliest epoch when voluntary exit can be processed
	{"validator_index", ValidatorIndexType},
}

var SignedVoluntaryExitType = &ContainerType{
	{"message", VoluntaryExitType},
	{"signature", BLSSignatureType},
}

func ProcessVoluntaryExit(input VoluntaryExitProcessInput, signedExit *SignedVoluntaryExit) error {
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
	return input.InitiateValidatorExit(currentEpoch, exit.ValidatorIndex)
}
