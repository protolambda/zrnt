package exits

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type VoluntaryExitFeature struct {
	Meta interface {
		meta.VersioningMeta
		meta.RegistrySizeMeta
		meta.ValidatorMeta
		meta.ExitMeta
	}
}

func (f *VoluntaryExitFeature) ProcessVoluntaryExits(ops []VoluntaryExit) error {
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
	Signature      BLSSignature
}

func (f *VoluntaryExitFeature) ProcessVoluntaryExit(exit *VoluntaryExit) error {
	currentEpoch := f.Meta.CurrentEpoch()
	if !f.Meta.IsValidIndex(exit.ValidatorIndex) {
		return errors.New("invalid exit validator index")
	}
	validator := f.Meta.Validator(exit.ValidatorIndex)
	// Verify that the validator is active
	if !validator.IsActive(currentEpoch) {
		return errors.New("validator must be active to be able to voluntarily exit")
	}
	// Verify the validator has not yet exited
	if validator.ExitEpoch != FAR_FUTURE_EPOCH {
		return errors.New("validator already exited")
	}
	// Exits must specify an epoch when they become valid; they are not valid before then
	if currentEpoch < exit.Epoch {
		return errors.New("invalid exit epoch")
	}
	// Verify the validator has been active long enough
	if currentEpoch < validator.ActivationEpoch+PERSISTENT_COMMITTEE_PERIOD {
		return errors.New("exit is too soon")
	}
	// Verify signature
	if !bls.BlsVerify(
		validator.Pubkey,
		ssz.SigningRoot(exit, VoluntaryExitSSZ),
		exit.Signature,
		f.Meta.GetDomain(DOMAIN_VOLUNTARY_EXIT, exit.Epoch)) {
		return errors.New("voluntary exit signature could not be verified")
	}
	// Initiate exit
	f.Meta.InitiateValidatorExit(f.Meta.CurrentEpoch(), exit.ValidatorIndex)
	return nil
}
