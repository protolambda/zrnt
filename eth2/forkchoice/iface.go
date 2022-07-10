package forkchoice

import (
	"context"

	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type Root = common.Root
type Epoch = common.Epoch
type Slot = common.Slot
type ValidatorIndex = common.ValidatorIndex
type Gwei = common.Gwei
type Checkpoint = common.Checkpoint
type NodeRef = common.NodeRef
type ExtendedNodeRef = common.ExtendedNodeRef
type SignedGwei int64
type NodeIndex uint64

type ForkchoiceView interface {
	CanonicalChain(anchorRoot Root, anchorSlot Slot) ([]ExtendedNodeRef, error)
	ClosestToSlot(anchor Root, slot Slot) (closest NodeRef, err error)
	CanonAtSlot(anchor Root, slot Slot, withBlock bool) (at NodeRef, err error)
	GetSlot(blockRoot Root) (slot Slot, ok bool)
	FindHead(anchorRoot Root, anchorSlot Slot) (NodeRef, error)
	InSubtree(anchor Root, root Root) (unknown bool, inSubtree bool)
	Search(anchor NodeRef, parentRoot *Root, slot *Slot) (nonCanon []NodeRef, canon []NodeRef, err error)
}

type ForkchoiceNodeInput interface {
	ProcessSlot(parent Root, slot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch)
	ProcessBlock(parent Root, blockRoot Root, blockSlot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) (ok bool)
}

type ForkchoiceGraph interface {
	ForkchoiceView
	ForkchoiceNodeInput
	Indices() map[NodeRef]NodeIndex
	ApplyScoreChanges(deltas []SignedGwei, justifiedEpoch Epoch, finalizedEpoch Epoch) error
	OnPrune(ctx context.Context, anchorRoot Root, anchorSlot Slot) error
}

type VoteInput interface {
	// ProcessAttestation overrides any previous vote, and applies voting weight to the new root/slot.
	// If the root/slot combination does not exist, no changes are made, and ok=false is returned.
	// It is up to the caller if nodes should be added, to then process the attestation.
	ProcessAttestation(index ValidatorIndex, blockRoot Root, headSlot Slot) (ok bool)
}

type VoteStore interface {
	VoteInput
	HasChanges() bool
	ComputeDeltas(indices map[NodeRef]NodeIndex, oldBalances []Gwei, newBalances []Gwei) []SignedGwei
}

type Forkchoice interface {
	ForkchoiceView
	ForkchoiceNodeInput
	VoteInput
	UpdateJustified(ctx context.Context, trigger Root, justified Checkpoint, finalized Checkpoint,
		justifiedStateBalances func() ([]Gwei, error)) error
	Pin() *NodeRef
	SetPin(root Root, slot Slot) error
	Justified() Checkpoint
	Finalized() Checkpoint
	Head() (NodeRef, error)
}
