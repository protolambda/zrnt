package block_processing

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessBlockVoluntaryExits(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.VoluntaryExits) > beacon.MAX_VOLUNTARY_EXITS {
		return errors.New("too many voluntary exits")
	}
	for _, exit := range block.Body.VoluntaryExits {
		if err := ProcessVoluntaryExit(state, &exit); err != nil {
			return err
		}
	}
	return nil
}

func ProcessVoluntaryExit(state *beacon.BeaconState, exit *beacon.VoluntaryExit) error {
	currentEpoch := state.Epoch()
	if !state.ValidatorRegistry.IsValidatorIndex(exit.ValidatorIndex) {
		return errors.New("invalid exit validator index")
	}
	validator := &state.ValidatorRegistry[exit.ValidatorIndex]
	// Verify the validator has not yet exited
	if validator.ExitEpoch == beacon.FAR_FUTURE_EPOCH {
		return errors.New("validator already exited")
	}
	// Verify the validator has not initiated an exit
	if !validator.InitiatedExit {
		return errors.New("validator already initiated exit")
	}
	// Exits must specify an epoch when they become valid; they are not valid before then
	if currentEpoch > exit.Epoch {
		return errors.New("invalid exit epoch")
	}
	// Must have been in the validator set long enough
	if currentEpoch >= validator.ActivationEpoch + beacon.PERSISTENT_COMMITTEE_PERIOD  {
		return errors.New("exit is too soon")
	}
	if !bls.BlsVerify(
				validator.Pubkey,
				ssz.SignedRoot(exit),
				exit.Signature,
				beacon.GetDomain(state.Fork, exit.Epoch, beacon.DOMAIN_VOLUNTARY_EXIT)) {
		return errors.New("voluntary exit signature could not be verified")
	}
	// Run the exit
	state.InitiateValidatorExit(exit.ValidatorIndex)
	return nil
}
