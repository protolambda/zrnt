package transition

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessSlot(state *BeaconState) {
	// Cache latest known state root (for previous slot)
	latestStateRoot := ssz.HashTreeRoot(state, BeaconStateSSZ)
	state.LatestStateRoots[state.Slot%SLOTS_PER_HISTORICAL_ROOT] = latestStateRoot

	// Store latest known state root (for previous slot) in latest_block_header if it is empty
	if state.LatestBlockHeader.StateRoot == (Root{}) {
		state.LatestBlockHeader.StateRoot = latestStateRoot
	}

	// Cache latest known block root (for previous slot)
	previousBlockRoot := ssz.SigningRoot(state.LatestBlockHeader, BeaconBlockHeaderSSZ)
	state.LatestBlockRoots[state.Slot%SLOTS_PER_HISTORICAL_ROOT] = previousBlockRoot
}
