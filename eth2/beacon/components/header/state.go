package header

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type BlockHeaderState struct {
	LatestBlockHeader BeaconBlockHeader
}

// Signing root of latest_block_header
func (state *BlockHeaderState) GetLatestBlockRoot() Root {
	return ssz.SigningRoot(state.LatestBlockHeader, BeaconBlockHeaderSSZ)
}

func (state *BlockHeaderState) StoreHeaderData(slot Slot, parent Root, body Root) {
	state.LatestBlockHeader = BeaconBlockHeader{
		Slot:       slot,
		ParentRoot: parent,
		// state_root: zeroed, overwritten with UpdateStateRoot(), once the post state is available.
		BodyRoot: body,
		// signature is always zeroed
	}
}

func (state *BlockHeaderState) UpdateLatestBlockRoot(stateRoot Root) Root {
	// Store latest known state root (for previous slot) in latest_block_header if it is empty
	if state.LatestBlockHeader.StateRoot == (Root{}) {
		state.LatestBlockHeader.StateRoot = stateRoot
	}
	return ssz.SigningRoot(state.LatestBlockHeader, BeaconBlockHeaderSSZ)
}

func (state *BlockHeaderState) UpdateStateRoot(root Root) {
	state.LatestBlockHeader.StateRoot = root
}
