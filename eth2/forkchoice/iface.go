package forkchoice

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon"
)

type Root = beacon.Root
type Epoch = beacon.Epoch
type Slot = beacon.Slot
type ValidatorIndex = beacon.ValidatorIndex
type Gwei = beacon.Gwei
type Checkpoint = beacon.Checkpoint
type NodeRef = beacon.NodeRef
type ExtendedNodeRef = beacon.ExtendedNodeRef
type SignedGwei int64
type NodeIndex uint64

type ForkchoiceView interface {
	CanonicalChain(anchorRoot Root, anchorSlot Slot) ([]ExtendedNodeRef, error)
	ClosestToSlot(anchor Root, slot Slot) (closest NodeRef, err error)
	CanonAtSlot(anchor Root, slot Slot, withBlock bool) (at NodeRef, err error)
	GetSlot(blockRoot Root) (slot Slot, ok bool)
	FindHead(anchorRoot Root, anchorSlot Slot) (NodeRef, error)
	IsAncestor(root Root, ofRoot Root) (unknown bool, isAncestor bool)
}

type ForkchoiceNodeInput interface {
	ProcessSlot(parent Root, slot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch)
	ProcessBlock(parent Root, blockRoot Root, blockSlot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch)
}

type ForkchoiceGraph interface {
	ForkchoiceView
	ForkchoiceNodeInput
	Indices() map[NodeRef]NodeIndex
	ApplyScoreChanges(deltas []SignedGwei, justifiedEpoch Epoch, finalizedEpoch Epoch) error
	OnPrune(ctx context.Context, anchorRoot Root, anchorSlot Slot) error
}

type VoteInput interface {
	ProcessAttestation(index ValidatorIndex, blockRoot Root, headSlot Slot)
}

type VoteStore interface {
	VoteInput
	ComputeDeltas(indices map[NodeRef]NodeIndex, oldBalances []Gwei, newBalances []Gwei) []SignedGwei
}

type Forkchoice interface {
	ForkchoiceView
	ForkchoiceNodeInput
	VoteInput
	Justified() Checkpoint
	Finalized() Checkpoint
	Head() (NodeRef, error)
}
