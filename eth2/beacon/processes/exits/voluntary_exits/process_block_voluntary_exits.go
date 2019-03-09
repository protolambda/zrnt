package voluntary_exits

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
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
	if !(
		validator.ExitEpoch > currentEpoch.GetDelayedActivationExitEpoch() &&
			currentEpoch > exit.Epoch &&
			bls.BlsVerify(
				validator.Pubkey,
				ssz.SignedRoot(exit),
				exit.Signature,
				beacon.GetDomain(state.Fork, exit.Epoch, beacon.DOMAIN_VOLUNTARY_EXIT))) {
		return errors.New("voluntary exit could not be verified")
	}
	state.InitiateValidatorExit(exit.ValidatorIndex)
	return nil
}
