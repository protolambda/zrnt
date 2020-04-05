package beacon

import (

	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

var BeaconBlockHeaderSSZ = zssz.GetSSZ((*BeaconBlockHeader)(nil))

type BeaconBlockHeader struct {
	Slot          Slot
	ProposerIndex ValidatorIndex
	ParentRoot    Root
	StateRoot     Root
	BodyRoot      Root
}

func (b *BeaconBlockHeader) HashTreeRoot() Root {
	return ssz.HashTreeRoot(b, BeaconBlockHeaderSSZ)
}

type SignedBeaconBlockHeader struct {
	Message   BeaconBlockHeader
	Signature BLSSignature
}
