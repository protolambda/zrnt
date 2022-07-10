package beacon

import (
	"context"

	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type ChainEntry interface {
	// Step of this entry
	Step() common.Step
	// BlockRoot returns the last block root, replicating the previous block root if the current slot has none.
	// There is only 1 block root, double block proposals by the same validator are accepted,
	// only the first is incorporated into the chain.
	BlockRoot() (root common.Root, err error)
	// The parent block root. If this is an empty slot, it will just be previous block root. Can also be zeroed if unknown.
	ParentRoot() (root common.Root, err error)
	// State root of the post-state of this entry, with or without block, depending on IsEmpty.
	// Should match state-root in the block at the same slot (if any)
	StateRoot() (common.Root, error)
	// The context of this chain entry (shuffling, proposers, etc.)
	EpochsContext(ctx context.Context) (*common.EpochsContext, error)
	// StateExclBlock retrieves the state of this slot.
	// - If IsEmpty: it is the state after processing slots to Slot() (incl.),
	//   with ProcessSlots(slot), but without any block processing.
	// - if not IsEmpty: post-block processing (if any block), excl. latest-header update of next slot.
	State(ctx context.Context) (common.BeaconState, error)
}

type SearchEntry struct {
	ChainEntry
	Canonical bool
}

type GenesisInfo struct {
	Time           common.Timestamp
	ValidatorsRoot common.Root
}

type Chain interface {
	// Get the chain entry for the given state root (post slot processing or post block processing)
	ByStateRoot(root common.Root) (entry ChainEntry, ok bool)
	// Get the chain entry for the given block root
	ByBlock(root common.Root) (entry ChainEntry, ok bool)
	// Get the chain entry for the given block root and slot, may be an empty slot,
	// or may be in-between slot processing and block processing if the parent block root is requested for the slot.
	ByBlockSlot(root common.Root, slot common.Slot) (entry ChainEntry, ok bool)
	// Get the blocks(s) with the given parent-root and/or slot.
	// Return all possible heads by default (if options are nil).
	Search(parentRoot *common.Root, slot *common.Slot) ([]SearchEntry, error)
	// Find closest ref in subtree, up to given slot (may return entry of fromBlockRoot itself),
	// without any blocks after fromBlockRoot.
	// Err if no entry, even not fromBlockRoot, could be found.
	Closest(fromBlockRoot common.Root, toSlot common.Slot) (entry ChainEntry, ok bool)
	// Returns true if the given root is something that builds (maybe indirectly) on the anchor.
	// I.e. if root is in the subtree of anchor.
	// If root == anchor, then it is also considered to be in the subtree here.
	InSubtree(anchor common.Root, root common.Root) (unknown bool, inSubtree bool)
	// Get the canonical entry at the given slot. Return nil if there is no block but the slot node exists.
	ByCanonStep(step common.Step) (entry ChainEntry, ok bool)
	Iter() (ChainIter, error)
	JustifiedCheckpoint() common.Checkpoint
	FinalizedCheckpoint() common.Checkpoint
	Justified() (ChainEntry, error)
	Finalized() (ChainEntry, error)
	Head() (ChainEntry, error)
	// First gets the closets ref from the given block root to the requested slot,
	// then transitions empty slots to get up to the requested slot.
	// A strict context should be provided to avoid costly long transitions.
	// An error is also returned if the fromBlockRoot is past the requested toSlot.
	Towards(ctx context.Context, fromBlockRoot common.Root, toSlot common.Slot) (ChainEntry, error)
	Genesis() GenesisInfo
}

type ChainIter interface {
	// Start is the minimum to reach to, inclusive. The step may exclude pre-block processing.
	Start() common.Step
	// End is the maximum to reach to, exclusive. The step may exclude post-block processing.
	End() common.Step
	// Entry fetches the chain entry at the given slot.
	// If the slot has no block but step.Block is true, then entry == nil, err == nil.
	// If the request is out of bounds or fails, an error is returned.
	// The step.Block on Start() and End() counts as bounds: chains may only store part of the slot.
	Entry(step common.Step) (entry ChainEntry, err error)
}
