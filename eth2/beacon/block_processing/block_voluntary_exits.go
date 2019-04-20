package block_processing

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessBlockVoluntaryExits(state *BeaconState, block *BeaconBlock) error {
	if len(block.Body.VoluntaryExits) > MAX_VOLUNTARY_EXITS {
		return errors.New("too many voluntary exits")
	}
	for _, exit := range block.Body.VoluntaryExits {
		if err := ProcessVoluntaryExit(state, &exit); err != nil {
			return err
		}
	}
	return nil
}

func ProcessVoluntaryExit(state *BeaconState, exit *VoluntaryExit) error {
	currentEpoch := state.Epoch()
	if !state.ValidatorRegistry.IsValidatorIndex(exit.ValidatorIndex) {
		return errors.New("invalid exit validator index")
	}
	validator := state.ValidatorRegistry[exit.ValidatorIndex]
	// Verify that the validator is active
	if !validator.IsActive(currentEpoch) {
		return errors.New("validator must be active to be able to voluntarily exit")
	}
	// Verify the validator has not yet exited
	if validator.ExitEpoch == FAR_FUTURE_EPOCH {
		return errors.New("validator already exited")
	}
	// Exits must specify an epoch when they become valid; they are not valid before then
	if currentEpoch > exit.Epoch {
		return errors.New("invalid exit epoch")
	}
	// Verify the validator has been active long enough
	if currentEpoch >= validator.ActivationEpoch + PERSISTENT_COMMITTEE_PERIOD  {
		return errors.New("exit is too soon")
	}
	if !bls.BlsVerify(
				validator.Pubkey,
				ssz.SigningRoot(exit),
				exit.Signature,
				GetDomain(state.Fork, exit.Epoch, DOMAIN_VOLUNTARY_EXIT)) {
		return errors.New("voluntary exit signature could not be verified")
	}
	// Initiate exit
	state.InitiateValidatorExit(exit.ValidatorIndex)
	return nil
}
