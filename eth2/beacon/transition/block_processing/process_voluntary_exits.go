package block_processing

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/transition"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessVoluntaryExits(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Voluntary_exits) > eth2.MAX_VOLUNTARY_EXITS {
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
	validator := &state.Validator_registry[exit.Validator_index]
	if !(validator.Exit_epoch > transition.Get_delayed_activation_exit_epoch(state.Epoch()) &&
		state.Epoch() > exit.Epoch &&
		bls.Bls_verify(validator.Pubkey, ssz.Signed_root(exit),
			exit.Signature, transition.Get_domain(state.Fork, exit.Epoch, eth2.DOMAIN_VOLUNTARY_EXIT))) {
		return errors.New("voluntary exit could not be verified")
	}
	transition.Initiate_validator_exit(state, exit.Validator_index)
	return nil
}
