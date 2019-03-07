package randao

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/hash"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessBlockRandao(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	propIndex := state.Get_beacon_proposer_index(state.Slot, false)
	proposer := &state.Validator_registry[propIndex]
	if !bls.Bls_verify(
		proposer.Pubkey,
		ssz.Hash_tree_root(state.Epoch()),
		block.Body.Randao_reveal,
		beacon.Get_domain(state.Fork, state.Epoch(), beacon.DOMAIN_RANDAO),
	) {
		return errors.New("randao invalid")
	}
	state.Latest_randao_mixes[state.Epoch()%beacon.LATEST_RANDAO_MIXES_LENGTH] = hash.XorBytes32(
		state.Get_randao_mix(state.Epoch()),
		hash.Hash(block.Body.Randao_reveal[:]))
	return nil
}
