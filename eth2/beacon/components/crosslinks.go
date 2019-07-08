package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz"
)

var CrosslinkSSZ = zssz.GetSSZ((*Crosslink)(nil))

type Crosslink struct {
	// Shard number
	Shard Shard
	// Crosslinking data from epochs [start....end-1]
	StartEpoch Epoch
	EndEpoch   Epoch
	// Root of the previous crosslink
	ParentRoot Root
	// Root of the crosslinked shard data since the previous crosslink
	DataRoot Root
}

type CrosslinksState struct {
	CurrentCrosslinks  [SHARD_COUNT]Crosslink
	PreviousCrosslinks [SHARD_COUNT]Crosslink
}
