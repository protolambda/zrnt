package operations

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type VoluntaryExits []VoluntaryExit

func (_ *VoluntaryExits) Limit() uint32 {
	return MAX_VOLUNTARY_EXITS
}

func (ops VoluntaryExits) Process(state *BeaconState) error {
	for _, op := range ops {
		if err := op.Process(state); err != nil {
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

func (exit *VoluntaryExit) Process(state *BeaconState) error {
	currentEpoch := state.Epoch()
	if !state.Validators.IsValidatorIndex(exit.ValidatorIndex) {
		return errors.New("invalid exit validator index")
	}
	validator := state.Validators[exit.ValidatorIndex]
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
	if !bls.BlsVerify(
		validator.Pubkey,
		ssz.SigningRoot(exit, VoluntaryExitSSZ),
		exit.Signature,
		state.GetDomain(DOMAIN_VOLUNTARY_EXIT, exit.Epoch)) {
		return errors.New("voluntary exit signature could not be verified")
	}
	// Initiate exit
	state.InitiateValidatorExit(exit.ValidatorIndex)
	return nil
}
