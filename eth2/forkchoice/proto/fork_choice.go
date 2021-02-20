package proto

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/forkchoice"
)

func NewProtoForkChoice(spec *beacon.Spec, finalized Checkpoint, justified Checkpoint,
	anchorRoot Root, anchorSlot Slot, anchorParent Root,
	initialBalances []Gwei, sink NodeSink) Forkchoice {
	return NewForkChoice(spec, finalized, justified, anchorRoot, anchorSlot, anchorParent,
		NewProtoArray(justified.Epoch, finalized.Epoch, sink),
		NewProtoVoteStore(spec), initialBalances)
}
