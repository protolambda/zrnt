package block

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz"
)

var BeaconBlockBodySSZ = zssz.GetSSZ((*BeaconBlockBody)(nil))

type BeaconBlockBody struct {
	RandaoRevealBlockData
	Eth1BlockData
	Graffiti Root

	BlockOperations
}

func (body *BeaconBlockBody) Process(state *BeaconState) error {
	if err := body.RandaoRevealBlockData.Process(state); err != nil {
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
