package block_header

import (
	"errors"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/bls"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessBlockHeader(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	// Verify that the slots match
	if block.Slot != state.Slot {
		return errors.New("slot of block does not match slot of state")
	}
	// Verify that the parent matches
	if block.PreviousBlockRoot != ssz.Hash_tree_root(state.LatestBlockHeader) {
		return errors.New("previous block root does not match root from latest state block header")
	}
	// Save current block as the new latest block
	state.LatestBlockHeader = block.GetTemporaryBlockHeader()

	propIndex := state.Get_beacon_proposer_index(state.Slot, false)
	// Verify proposer signature
	proposer := &state.Validator_registry[propIndex]
	// Block signature
	if !bls.Bls_verify(
		proposer.Pubkey,
		ssz.Signed_root(block),
		block.Signature,
		beacon.Get_domain(state.Fork, state.Epoch(), beacon.DOMAIN_BEACON_BLOCK)) {
		return errors.New("block signature invalid")
	}

}
