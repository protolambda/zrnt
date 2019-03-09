package transition

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

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
