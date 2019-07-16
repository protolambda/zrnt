package block

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

var BeaconBlockSSZ = zssz.GetSSZ((*BeaconBlock)(nil))

type BeaconBlock struct {
	Slot       Slot
	ParentRoot Root
	StateRoot  Root
	Body       BeaconBlockBody
	Signature  BLSSignature
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
