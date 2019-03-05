package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/merkle"
)

// Let previous_block_root be the hash_tree_root of the previous beacon block processed in the chain.
func SlotTransition(state *beacon.BeaconState, previous_block_root eth2.Root) {
	state.Slot += 1

	state.Latest_block_roots[(state.Slot-1)%eth2.LATEST_BLOCK_ROOTS_LENGTH] = previous_block_root
	if state.Slot%eth2.LATEST_BLOCK_ROOTS_LENGTH == 0 {
		// yes, this is ugly, typing requires us to be explict when we want to merkleize a list of non-bytes32 items.
		merkle_input := make([]eth2.Bytes32, len(state.Latest_block_roots))
		for i := 0; i < len(state.Latest_block_roots); i++ {
			merkle_input[i] = eth2.Bytes32(state.Latest_block_roots[i])
		}
		state.Batched_block_roots = append(state.Batched_block_roots, merkle.Merkle_root(merkle_input))
	}
}
