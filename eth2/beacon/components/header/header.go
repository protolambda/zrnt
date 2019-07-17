package header

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz"
)

var BeaconBlockHeaderSSZ = zssz.GetSSZ((*BeaconBlockHeader)(nil))

type BeaconBlockHeader struct {
	Slot       Slot
	ParentRoot Root
	StateRoot  Root
	BodyRoot   Root // Where the body would be, just a root embedded here.
	Signature  BLSSignature
}
