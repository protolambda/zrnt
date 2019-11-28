package header

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	. "github.com/protolambda/ztyp/view"
)

type LatestBlockHeader struct{ *ContainerView }

func (v *LatestBlockHeader) GetStateRoot() (Root, error) {

}

// Signing root of latest_block_header
func (v *LatestBlockHeader) GetLatestBlockRoot() Root {
	return v.ViewRoot(h)
}

func (v *LatestBlockHeader) UpdateLatestBlockStateRoot(stateRoot Root) {
	// Store latest known state root (for previous slot) in latest_block_header if it is empty
	if state.LatestBlockHeader.StateRoot == (Root{}) {
		state.LatestBlockHeader.StateRoot = stateRoot
	}
}

func (v *LatestBlockHeader) UpdateStateRoot(root Root) {
	state.LatestBlockHeader.StateRoot = root
}
