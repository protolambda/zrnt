package block

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz"
)

var BeaconBlockBodySSZ = zssz.GetSSZ((*BeaconBlockBody)(nil))

type BeaconBlockBody struct {
	RandaoBlockData
	Eth1BlockData // Eth1 data vote
	Graffiti Root // Arbitrary data

	BlockOperations
}

func (body *BeaconBlockBody) Process(state *BeaconState) error {
	if err := body.RandaoBlockData.Process(state); err != nil {
		return nil
	}
	if err := body.Eth1BlockData.Process(state); err != nil {
		return nil
	}
	if err := body.BlockOperations.Process(state); err != nil {
		return nil
	}
	return nil
}
