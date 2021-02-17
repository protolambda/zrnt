package chain

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/db/states"
	"sync"
)

type Root = beacon.Root
type Epoch = beacon.Epoch
type Slot = beacon.Slot
type ValidatorIndex = beacon.ValidatorIndex
type Gwei = beacon.Gwei
type Checkpoint = beacon.Checkpoint

// Step combines a Slot and bool for block processing being included or not.
type Step uint64

func AsStep(slot Slot, block bool) Step {
	if slot&(1<<63) != 0 {
		panic("slot overflow")
	}
	out := Step(slot) << 1
	if block {
		out++
	}
	return out
}

func (st Step) String() string {
	if st.Block() {
		return fmt.Sprintf("%d:1", st.Slot())
	} else {
		return fmt.Sprintf("%d:0", st.Slot())
	}
}

func (st Step) Slot() Slot {
	return Slot(st >> 1)
}

func (st Step) Block() bool {
	return st&1 != 0
}

type ChainEntry interface {
	// Step of this entry
	Step() Step
	// BlockRoot returns the last block root, replicating the previous block root if the current slot has none.
	// There is only 1 block root, double block proposals by the same validator are accepted,
	// only the first is incorporated into the chain.
	BlockRoot() (root Root)
	// The parent block root. If this is an empty slot, it will just be previous block root. Can also be zeroed if unknown.
	ParentRoot() (root Root)
	// State root of the post-state of this entry, with or without block, depending on IsEmpty.
	// Should match state-root in the block at the same slot (if any)
	StateRoot() Root
	// The context of this chain entry (shuffling, proposers, etc.)
	EpochsContext(ctx context.Context) (*beacon.EpochsContext, error)
	// StateExclBlock retrieves the state of this slot.
	// - If IsEmpty: it is the state after processing slots to Slot() (incl.),
	//   with ProcessSlots(slot), but without any block processing.
	// - if not IsEmpty: post-block processing (if any block), excl. latest-header update of next slot.
	State(ctx context.Context) (*beacon.BeaconStateView, error)
}

type Chain interface {
	// Get the chain entry for the given state root (post slot processing or post block processing)
	ByStateRoot(root Root) (entry ChainEntry, ok bool)
	// Get the chain entry for the given block root
	ByBlock(root Root) (entry ChainEntry, ok bool)
	// Get the chain entry for the given block root and slot, may be an empty slot,
	// or may be in-between slot processing and block processing if the parent block root is requested for the slot.
	ByBlockSlot(root Root, slot Slot) (entry ChainEntry, ok bool)
	// Find closest ref in subtree, up to given slot (may return entry of fromBlockRoot itself),
	// without any blocks after fromBlockRoot.
	// Err if no entry, even not fromBlockRoot, could be found.
	Closest(fromBlockRoot Root, toSlot Slot) (entry ChainEntry, ok bool)
	// Returns true if the given root is something that builds (maybe indirectly)
	// on the ofRoot on the same chain.
	// If root == ofRoot, then it is NOT considered an ancestor here.
	IsAncestor(root Root, ofRoot Root) (unknown bool, isAncestor bool)
	// Get the canonical entry at the given slot. Return nil if there is no block but the slot node exists.
	ByCanonStep(step Step) (entry ChainEntry, ok bool)
	Iter() (ChainIter, error)
}

type ChainIter interface {
	// Start is the minimum to reach to, inclusive. The step may exclude pre-block processing.
	Start() Step
	// End is the maximum to reach to, exclusive. The step may exclude post-block processing.
	End() Step
	// Entry fetches the chain entry at the given slot.
	// If the slot has no block but step.Block is true, then entry == nil, err == nil.
	// If the request is out of bounds or fails, an error is returned.
	// The step.Block on Start() and End() counts as bounds: chains may only store part of the slot.
	Entry(step Step) (entry ChainEntry, err error)
}

type BlockSlotKey struct {
	Slot Slot
	Root Root
}

type FullChain interface {
	Chain
	HotChain
	ColdChain
}

type HotColdChain struct {
	// sync.Mutex to control access to the hot and cold chain at the same time.
	// The HotChain is allowed to move data to the cold chain, but not reverse.
	// Internally it is safe to query the hot chain first, and the cold chain later.
	// The ColdChain is only allowed to remove data.
	sync.Mutex
	HotChain
	ColdChain
	Spec *beacon.Spec
}

var _ FullChain = (*HotColdChain)(nil)

func NewHotColdChain(anchorState *beacon.BeaconStateView, spec *beacon.Spec, stateDB states.DB) (*HotColdChain, error) {
	c := &HotColdChain{
		HotChain:  nil,
		ColdChain: NewFinalizedChain(spec, stateDB),
		Spec:      spec,
	}
	hotCh, err := NewUnfinalizedChain(anchorState, BlockSinkFn(c.hotToCold), spec)
	if err != nil {
		return nil, err
	}
	c.HotChain = hotCh

	return c, nil
}

func (hc *HotColdChain) hotToCold(ctx context.Context, entry ChainEntry, canonical bool) error {
	if canonical {
		return hc.ColdChain.OnFinalizedEntry(ctx, entry)
	}
	// TODO keep track of pruned non-finalized blocks?
	return nil
}

func (hc *HotColdChain) ByStateRoot(root Root) (entry ChainEntry, ok bool) {
	hc.Lock()
	defer hc.Unlock()
	entry, ok = hc.HotChain.ByStateRoot(root)
	if ok {
		return entry, ok
	}
	return hc.ColdChain.ByStateRoot(root)
}

func (hc *HotColdChain) ByBlock(root Root) (entry ChainEntry, ok bool) {
	hc.Lock()
	defer hc.Unlock()
	entry, ok = hc.HotChain.ByBlock(root)
	if ok {
		return entry, ok
	}
	return hc.ColdChain.ByBlock(root)
}

func (hc *HotColdChain) ByBlockSlot(root Root, slot Slot) (entry ChainEntry, ok bool) {
	hc.Lock()
	defer hc.Unlock()
	entry, ok = hc.HotChain.ByBlockSlot(root, slot)
	if ok {
		return entry, ok
	}
	return hc.ColdChain.ByBlockSlot(root, slot)
}

func (hc *HotColdChain) Closest(fromBlockRoot Root, toSlot Slot) (entry ChainEntry, ok bool) {
	hc.Lock()
	defer hc.Unlock()
	entry, ok = hc.HotChain.Closest(fromBlockRoot, toSlot)
	if ok {
		return entry, ok
	}
	return hc.ColdChain.Closest(fromBlockRoot, toSlot)
}

func (hc *HotColdChain) IsAncestor(root Root, ofRoot Root) (unknown bool, isAncestor bool) {
	hc.Lock()
	defer hc.Unlock()

	// Tricky, but follow hot-to-cold to avoid missing data when it moves from hot to cold while processing.

	// can't be ancestors if they are equal.
	if root == ofRoot {
		return false, false
	}
	// if the first of the two roots is known in the hot chain, just have the hot chain deal with it.
	unknown, isAncestor = hc.HotChain.IsAncestor(root, ofRoot)
	if !unknown {
		return false, isAncestor
	}
	fin := hc.HotChain.Finalized()
	unknown, isAncestor = hc.HotChain.IsAncestor(root, fin.Root)
	if !unknown {
		// The root is in the hot subtree, now make sure the other root exists in the cold chain
		_, ok := hc.ColdChain.ByBlock(ofRoot)
		return !ok, ok
	}

	// Both are not in the hot chain, have the hot chain deal with it.
	return hc.ColdChain.IsAncestor(root, ofRoot)
}

func (hc *HotColdChain) ByCanonStep(step Step) (entry ChainEntry, ok bool) {
	hc.Lock()
	defer hc.Unlock()
	entry, ok = hc.HotChain.ByCanonStep(step)
	if ok {
		return entry, ok
	}
	return hc.ColdChain.ByCanonStep(step)
}

func (hc *HotColdChain) Iter() (ChainIter, error) {
	hc.Lock()
	defer hc.Unlock()
	hotIt, err := hc.HotChain.Iter()
	if err != nil {
		return nil, fmt.Errorf("cannot iter hot part: %v", err)
	}
	coldIt, err := hc.ColdChain.Iter()
	if err != nil {
		return nil, fmt.Errorf("cannot iter cold part: %v", err)
	}
	return &FullChainIter{
		HotIter:  hotIt,
		ColdIter: coldIt,
	}, nil
}

type FullChainIter struct {
	HotIter  ChainIter
	ColdIter ChainIter
}

func (fi *FullChainIter) Start() Step {
	return fi.ColdIter.Start()
}

func (fi *FullChainIter) End() Step {
	return fi.HotIter.End()
}

func (fi *FullChainIter) Entry(step Step) (entry ChainEntry, err error) {
	if step < fi.ColdIter.End() {
		return fi.ColdIter.Entry(step)
	} else {
		return fi.HotIter.Entry(step)
	}
}

// TODO: chain copy
