package block

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

var BeaconBlockBodySSZ = zssz.GetSSZ((*BeaconBlockBody)(nil))

type BeaconBlockBody struct {
	RandaoReveal BLSSignature
	Eth1BlockData
	Graffiti Root

	BlockOperations
}

var BeaconBlockSSZ = zssz.GetSSZ((*BeaconBlock)(nil))

type BeaconBlock struct {
	Slot       Slot
	ParentRoot Root
	StateRoot  Root
	// Body
	Body BeaconBlockBody
	// Signature
	Signature BLSSignature
}

func (block *BeaconBlock) Header() *BeaconBlockHeader {
	return &BeaconBlockHeader{
		Slot:       block.Slot,
		ParentRoot: block.ParentRoot,
		StateRoot:  block.StateRoot,
		BodyRoot:   ssz.HashTreeRoot(block.Body, BeaconBlockBodySSZ),
		Signature:  block.Signature,
	}
}
