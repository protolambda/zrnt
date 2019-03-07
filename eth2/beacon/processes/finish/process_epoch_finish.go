package finish

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/merkle"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

func ProcessEpochFinish(state *beacon.BeaconState) {
	current_epoch := state.Epoch()
	next_epoch := current_epoch + 1
	// Set active index root
	index_root_position := (next_epoch+beacon.ACTIVATION_EXIT_DELAY)%beacon.LATEST_ACTIVE_INDEX_ROOTS_LENGTH
	state.Latest_active_index_roots[index_root_position] =
		ssz.Hash_tree_root(state.Validator_registry.Get_active_validator_indices(next_epoch+beacon.ACTIVATION_EXIT_DELAY))
	state.Latest_slashed_balances[next_epoch%beacon.LATEST_SLASHED_EXIT_LENGTH] =
		state.Latest_slashed_balances[current_epoch%beacon.LATEST_SLASHED_EXIT_LENGTH]

	// Set randao mix
	state.Latest_randao_mixes[next_epoch % beacon.LATEST_RANDAO_MIXES_LENGTH] = state.Get_randao_mix(current_epoch)
	// Set historical root accumulator
	if next_epoch % beacon.SLOTS_PER_HISTORICAL_ROOT.ToEpoch() == 0 {
		roots := make([]beacon.Bytes32, beacon.SLOTS_PER_HISTORICAL_ROOT * 2)
		for i := beacon.Slot(0); i < beacon.SLOTS_PER_HISTORICAL_ROOT; i++ {
			roots[i] = beacon.Bytes32(state.Latest_block_roots[i])
			roots[i+beacon.SLOTS_PER_HISTORICAL_ROOT] = beacon.Bytes32(state.Latest_state_roots[i])
		}
		// Merkleleize the roots into one root
		historicalRoot := merkle.Merkle_root(roots)
		state.HistoricalRoots = append(state.HistoricalRoots, historicalRoot)
	}
	// Rotate current/previous epoch attestations
	state.PreviousEpochAttestations = state.CurrentEpochAttestations
	state.CurrentEpochAttestations = make([]beacon.PendingAttestation, 0)
}
