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
	ProcessVoluntaryExits(ops []SignedVoluntaryExit) error
	ProcessVoluntaryExit(signedExit *SignedVoluntaryExit) error
}

type VoluntaryExitFeature struct {
	Meta interface {
		meta.ActiveIndices
		meta.Pubkeys
		meta.SigDomain
		meta.ExitEpoch
		meta.ActivationEpoch
		meta.Versioning
		meta.RegistrySize
		meta.Exits
	}
}

func (f *VoluntaryExitFeature) ProcessVoluntaryExits(ops []SignedVoluntaryExit) error {
	for i := range ops {
		if err := f.ProcessVoluntaryExit(&ops[i]); err != nil {
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

func (f *VoluntaryExitFeature) ProcessVoluntaryExit(signedExit *SignedVoluntaryExit) error {
	exit := &signedExit.Message
	currentEpoch, err := f.Meta.CurrentEpoch()
	if err != nil {
		return err
	}
	if valid, err := f.Meta.IsValidIndex(exit.ValidatorIndex); err != nil {
		return err
	} else if !valid {
		return errors.New("invalid exit validator index")
	}
	// Verify that the validator is active
	if isActive, err := f.Meta.IsActive(exit.ValidatorIndex, currentEpoch); err != nil {
		return err
	} else if !isActive {
		return errors.New("validator must be active to be able to voluntarily exit")
	}
	scheduledExitEpoch, err := f.Meta.ExitEpoch(exit.ValidatorIndex)
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
	registeredActivationEpoch, err := f.Meta.ActivationEpoch(exit.ValidatorIndex)
	if err != nil {
		return err
	}
	// Verify the validator has been active long enough
	if currentEpoch < registeredActivationEpoch+PERSISTENT_COMMITTEE_PERIOD {
		return errors.New("exit is too soon")
	}
	pubkey, err := f.Meta.Pubkey(exit.ValidatorIndex)
	if err != nil {
		return err
	}
	domain, err := f.Meta.GetDomain(DOMAIN_VOLUNTARY_EXIT, exit.Epoch)
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
	return f.Meta.InitiateValidatorExit(currentEpoch, exit.ValidatorIndex)
}
