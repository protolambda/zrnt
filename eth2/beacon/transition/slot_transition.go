package transition

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

// TODO state caching


func AdvanceSlot(state *beacon.BeaconState) {
	state.LatestStateRoots[state.Slot%beacon.SLOTS_PER_HISTORICAL_ROOT] = ssz.HashTreeRoot(state)
	state.Slot += 1
	if state.LatestBlockHeader.StateRoot == (beacon.Root{}) {
		// previous slot is safe, ignore error
		stRoot, _ := state.GetStateRoot(state.Slot - 1)
		state.LatestBlockHeader.StateRoot = stRoot
	}
	state.LatestBlockRoots[(state.Slot-1)%beacon.SLOTS_PER_HISTORICAL_ROOT] = ssz.HashTreeRoot(state.LatestBlockHeader)
}
