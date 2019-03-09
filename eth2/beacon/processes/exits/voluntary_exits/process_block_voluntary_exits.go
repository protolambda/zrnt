package voluntary_exits

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessBlockVoluntaryExits(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Voluntary_exits) > beacon.MAX_VOLUNTARY_EXITS {
		return errors.New("too many voluntary exits")
	}
	for _, exit := range block.Body.Voluntary_exits {
		if err := ProcessVoluntaryExit(state, &exit); err != nil {
			return err
		}
	}
	return nil
}

func ProcessVoluntaryExit(state *beacon.BeaconState, exit *beacon.VoluntaryExit) error {
	current_epoch := state.Epoch()
	if !state.Validator_registry.Is_validator_index(exit.Validator_index) {
		return errors.New("invalid exit validator index")
	}
	validator := &state.Validator_registry[exit.Validator_index]
	if !(
		validator.Exit_epoch > current_epoch.Get_delayed_activation_exit_epoch() &&
			current_epoch > exit.Epoch &&
			bls.Bls_verify(
				validator.Pubkey,
				ssz.Signed_root(exit),
				exit.Signature,
				beacon.Get_domain(state.Fork, exit.Epoch, beacon.DOMAIN_VOLUNTARY_EXIT))) {
		return errors.New("voluntary exit could not be verified")
	}
	state.Initiate_validator_exit(exit.Validator_index)
	return nil
}
