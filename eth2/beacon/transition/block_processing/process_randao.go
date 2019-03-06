package block_processing

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/transition"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/hash"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessRandao(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	propIndex := transition.Get_beacon_proposer_index(state, state.Slot, false)
	proposer := &state.Validator_registry[propIndex]
	if !bls.Bls_verify(
		proposer.Pubkey,
		ssz.Hash_tree_root(state.Epoch()),
		block.Body.Randao_reveal,
		transition.Get_domain(state.Fork, state.Epoch(), eth2.DOMAIN_RANDAO),
	) {
		return errors.New("randao invalid")
	}
	state.Latest_randao_mixes[state.Epoch()%eth2.LATEST_RANDAO_MIXES_LENGTH] = hash.XorBytes32(
		transition.Get_randao_mix(state, state.Epoch()),
		hash.Hash(block.Body.Randao_reveal[:]))
	return nil
}
