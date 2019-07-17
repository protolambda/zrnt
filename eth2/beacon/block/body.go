package block

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz"
)

var BeaconBlockBodySSZ = zssz.GetSSZ((*BeaconBlockBody)(nil))

type BeaconBlockBody struct {
	RandaoReveal BLSSignature
	Eth1Data Eth1Data // Eth1 data vote
	Graffiti Root     // Arbitrary data

	BlockOperations
}

func (body *BeaconBlockBody) Process(state *BeaconState) error {
	if err := state.ProcessRandaoReveal(state, body.RandaoReveal); err != nil {
		return nil
	}
	if err := state.ProcessEth1Vote(body.Eth1Data); err != nil {
		return nil
	}
	if err := body.BlockOperations.Process(state); err != nil {
		return nil
	}
	return nil
}
