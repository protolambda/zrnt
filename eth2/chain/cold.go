package chain

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/db/states"
	"sort"
	"sync"
)

type ColdChain interface {
	ColdStart() Step
	ColdEnd() Step
	OnFinalizedEntry(ctx context.Context, entry ChainEntry) error
	Chain
}

type FinalizedEntryView struct {
	step     Step
	finChain *FinalizedChain
}

func (e *FinalizedEntryView) Step() Step {
	return e.step
}

func (e *FinalizedEntryView) ParentRoot() (root Root) {
	return e.finChain.entryParentRoot(e.step.Slot())
}

func (e *FinalizedEntryView) BlockRoot() (root Root) {
	return e.finChain.entryBlockRoot(e.step)
}

func (e *FinalizedEntryView) StateRoot() Root {
	return e.finChain.entryStateRoot(e.step)
}

func (e *FinalizedEntryView) EpochsContext(ctx context.Context) (*phase0.EpochsContext, error) {
	return e.finChain.entryGetEpochsContext(ctx, e.step)
}

func (e *FinalizedEntryView) State(ctx context.Context) (*phase0.BeaconStateView, error) {
	return e.finChain.entryGetState(ctx, e.step)
}

// The finalized chain excludes the finalization node (i.e. node at start of finalized epoch),
// this can node can be found as anchor in the UnfinalizedChain instead.
// If this anchor node is filled with a block,
// it may happen that the node at the slot without the block is in the FinalizedChain,
// and the node at the same slot with the block is in the UnfinalizedChain.
type FinalizedChain struct {
	sync.RWMutex

	// Cache of pubkeys, may contain pubkeys that are not finalized,
	// but finalized state will not be in conflict with this cache.
	PubkeyCache *phase0.PubkeyCache

	// Block roots, starting at Anchor, indexed by step.
	// A blockRoot is a copy of the previous root if a slot is empty.
	// Every pre-block entry has a copy of the last block too.
	BlockRoots []Root

	// StateRoots are the state roots before and after applying a block, starting at Anchor, indexed by step index.
	StateRoots []Root

	// BlockRootsMap maps roots of BlockRoots back to slots (the first we know of, we may have pruned earlier slots)
	BlockRootsMap map[Root]Slot

	// StateRootsMaps maps roots of StateRoots back to steps
	StateRootsMap map[Root]Step

	// Spec is holds configuration information for the parameters and types of the chain
	Spec *common.Spec

	StateDB states.DB
}

var _ ColdChain = (*FinalizedChain)(nil)

func NewFinalizedChain(spec *common.Spec, stateDB states.DB) *FinalizedChain {
	initialCapacity := 200
	return &FinalizedChain{
		PubkeyCache:   phase0.EmptyPubkeyCache(),
		BlockRoots:    make([]Root, 0, initialCapacity),
		StateRoots:    make([]Root, 0, initialCapacity),
		BlockRootsMap: make(map[Root]Slot, initialCapacity),
		StateRootsMap: make(map[Root]Step, initialCapacity),
		Spec:          spec,
		StateDB:       stateDB,
	}
}

type ColdChainIter struct {
	Chain              Chain
	StartStep, EndStep Step
}

var _ ChainIter = (*ColdChainIter)(nil)

func (fi *ColdChainIter) Start() Step {
	return fi.StartStep
}

func (fi *ColdChainIter) End() Step {
	return fi.EndStep
}

func (fi *ColdChainIter) Entry(step Step) (entry ChainEntry, err error) {
	start := fi.Start()
	if step < start {
		return nil, fmt.Errorf("query too low")
	}
	end := fi.End()
	if step >= end {
		return nil, fmt.Errorf("query too high")
	}
	entry, ok := fi.Chain.ByCanonStep(step)
	if !ok {
		return nil, fmt.Errorf("failed to fetch %s from cold chain iter: %v", step, err)
	}
	return entry, nil
}

func (f *FinalizedChain) Iter() (ChainIter, error) {
	f.RLock()
	defer f.RUnlock()
	return &ColdChainIter{
		Chain:     f,
		StartStep: f.ColdStart(),
		EndStep:   f.ColdEnd(),
	}, nil
}

// Start of the cold chain (inclusive). Also see End
func (f *FinalizedChain) ColdStart() Step {
	f.RLock()
	defer f.RUnlock()
	return f.start()
}

func (f *FinalizedChain) start() Step {
	if len(f.StateRoots) == 0 {
		return 0
	}
	return f.StateRootsMap[f.StateRoots[0]]
}

// End step of the cold chain part (exclusive), should equal the epoch start slot of the finalized checkpoint,
// with or without block (depends on if it's a gap slot or not).
// The FinalizedChain is empty if Start() == End() == 0
func (f *FinalizedChain) ColdEnd() Step {
	f.RLock()
	defer f.RUnlock()
	return f.end()
}

func (f *FinalizedChain) end() Step {
	return f.start() + Step(len(f.StateRoots))
}

func (f *FinalizedChain) ByStateRoot(root Root) (entry ChainEntry, ok bool) {
	f.RLock()
	defer f.RUnlock()
	step, ok := f.StateRootsMap[root]
	if !ok {
		return nil, false
	}
	return f.byCanonStep(step)
}

func (f *FinalizedChain) ByBlock(root Root) (entry ChainEntry, ok bool) {
	f.RLock()
	defer f.RUnlock()
	slot, ok := f.BlockRootsMap[root]
	if !ok {
		return nil, false
	}
	return f.byBlockSlot(root, slot)
}

func (f *FinalizedChain) ByBlockSlot(root Root, slot Slot) (entry ChainEntry, ok bool) {
	f.RLock()
	defer f.RUnlock()
	return f.byBlockSlot(root, slot)
}

func (f *FinalizedChain) byBlockSlot(root Root, slot Slot) (entry ChainEntry, ok bool) {
	start, end := f.start().Slot(), f.end().Slot()
	if slot < start || slot > end {
		return nil, false
	}
	// check if block is known
	blockSlot, ok := f.BlockRootsMap[root]
	if !ok {
		return nil, false
	}
	// if block is older than the requested slot, we are not looking at the block itself,
	// but some gap slot after it.
	if blockSlot < slot {
		res, ok := f.byCanonStep(AsStep(slot, false))
		if !ok {
			return nil, false
		}
		// make sure the block root matches what we expect
		// (because checking the block root exists before the slot is not sufficient)
		if res.BlockRoot() != root {
			return nil, false
		}
		return res, true
	} else if blockSlot > slot {
		return nil, false
	} else {
		// block slot at root is same as requested, thus we are looking for a canon slot with block processed.
		return f.byCanonStep(AsStep(slot, true))
	}
}

func (f *FinalizedChain) Closest(fromBlockRoot Root, toSlot Slot) (entry ChainEntry, ok bool) {
	f.RLock()
	defer f.RUnlock()
	// check if block is known
	blockSlot, ok := f.BlockRootsMap[fromBlockRoot]
	if !ok {
		return nil, false
	}
	if blockSlot > toSlot {
		return nil, false
	}
	if blockSlot == toSlot {
		return f.byCanonStep(AsStep(toSlot, true))
	}
	// search in range: the block (incl), upto (excl) the block (if any) at the queried slot
	rangeStart := AsStep(blockSlot, true)
	rangeEnd := AsStep(toSlot, true)
	offset := int(rangeStart - f.start())
	n := int(rangeEnd - rangeStart)
	// find the first entry with non-equal block root (we know the first is equal)
	i := sort.Search(n, func(i int) bool {
		return f.BlockRoots[offset+i] != fromBlockRoot
	})
	foundStep := rangeStart + Step(i-1)
	return f.byCanonStep(foundStep)
}

func (f *FinalizedChain) Search(parentRoot *Root, slot *Slot) ([]SearchEntry, error) {
	f.RLock()
	defer f.RUnlock()
	// Searching the cold chain is a lot easier: there is at most 1 entry to retrieve.
	if slot != nil {
		entry, ok := f.ByCanonStep(AsStep(*slot, true))
		if !ok {
			return nil, nil
		}
		if parentRoot != nil && entry.ParentRoot() != *parentRoot {
			return nil, nil
		}
		return []SearchEntry{{ChainEntry: entry, Canonical: true}}, nil
	}
	if parentRoot != nil {
		slot, ok := f.BlockRootsMap[*parentRoot]
		if !ok {
			return nil, nil
		}
		parentStep := AsStep(slot, true)
		start := f.start()
		end := f.end()
		// steps of two, to always have a block entry, and skip over the pre-state of everything
		for i := parentStep + 2; i < end; i += 2 {
			// different root after the parent node? found our target then!
			if f.BlockRoots[i-start] != *parentRoot {
				entry, ok := f.byCanonStep(i)
				if !ok {
					return nil, nil
				}
				return []SearchEntry{{ChainEntry: entry, Canonical: true}}, nil
			}
		}
		return nil, nil
	}
	// cold chain has no heads, nothing to default to.
	return nil, nil
}

func (f *FinalizedChain) InSubtree(anchor Root, root Root) (unknown bool, inSubtree bool) {
	f.RLock()
	defer f.RUnlock()

	slot, ok := f.BlockRootsMap[root]
	if !ok {
		return true, false
	}
	anchorSlot, ok := f.BlockRootsMap[anchor]
	if !ok {
		return true, false
	}
	return false, anchorSlot <= slot
}

func (f *FinalizedChain) ByCanonStep(step Step) (entry ChainEntry, ok bool) {
	f.RLock()
	defer f.RUnlock()
	return f.byCanonStep(step)
}

func (f *FinalizedChain) byCanonStep(step Step) (entry ChainEntry, ok bool) {
	if start := f.ColdStart(); step < start {
		return nil, false
	}
	if end := f.ColdEnd(); step >= end {
		return nil, false
	}
	return &FinalizedEntryView{
		step:     step,
		finChain: f,
	}, true
}

func (f *FinalizedChain) OnFinalizedEntry(ctx context.Context, entry ChainEntry) error {
	f.Lock()
	defer f.Unlock()
	next := entry.Step()
	blockRoot := entry.BlockRoot()

	// If the chain is not empty, we need to verify consistency with what we add.
	if len(f.StateRoots) != 0 {
		end := f.ColdEnd()
		if end > next {
			return fmt.Errorf("received finalized entry %s at %s, but already finalized up to later step %s", blockRoot, next, end)
		}
		// Any direct follow-up (block after slot, slot after block),
		// or if the end is a gap slot, it may be left empty.
		if !(end+1 == next || (!end.Block() && end+2 == next)) {
			return fmt.Errorf("consistency issue, got %s and cannot append to %s", next, end)
		}
		// check parent root
		parent := entry.ParentRoot()
		if last := f.BlockRoots[len(f.BlockRoots)-1]; last != parent {
			return fmt.Errorf("parent root %s in block does not match last block root %s", parent, last)
		}
	}

	// Before modifying the chain tracking, try to store the state, so it is safe to abort on error
	state, err := entry.State(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve state of new finalized entry %s: %v", next, err)
	}
	if _, err := f.StateDB.Store(ctx, state); err != nil {
		return fmt.Errorf("failed to store state of new finalized entry %s: %v", next, err)
	}

	// Add block (may be a repeat of last)
	f.BlockRoots = append(f.BlockRoots, blockRoot)
	// Track the first occurrence
	if _, ok := f.BlockRootsMap[blockRoot]; !ok {
		f.BlockRootsMap[blockRoot] = next.Slot()
	}

	// Add new state
	stateRoot := entry.StateRoot()
	f.StateRoots = append(f.StateRoots, stateRoot)
	f.StateRootsMap[stateRoot] = next
	return nil
}

func (f *FinalizedChain) entryParentRoot(slot Slot) (root Root) {
	return f.entryBlockRoot(AsStep(slot, false))
}

func (f *FinalizedChain) entryBlockRoot(step Step) (root Root) {
	f.RLock()
	defer f.RUnlock()
	start := f.start()
	end := f.end()
	if start < step || step >= end {
		panic("out of bounds internal usage error")
	}
	return f.BlockRoots[step-start]
}

func (f *FinalizedChain) entryStateRoot(step Step) Root {
	f.RLock()
	defer f.RUnlock()
	return f.stateRoot(step)
}

func (f *FinalizedChain) stateRoot(step Step) Root {
	start := f.start()
	end := f.end()
	if start < step || step >= end {
		panic("out of bounds internal usage error")
	}
	return f.StateRoots[step-start]
}

func (f *FinalizedChain) entryGetEpochsContext(ctx context.Context, step Step) (*phase0.EpochsContext, error) {
	f.RLock()
	defer f.RUnlock()
	epc := &phase0.EpochsContext{
		PubkeyCache: f.PubkeyCache,
	}
	// We do not store shuffling for older epochs
	// TODO: maybe store it after all, for archive node functionality?
	state, err := f.getState(ctx, step)
	if err != nil {
		return nil, err
	}
	if err := epc.LoadShuffling(state); err != nil {
		return nil, err
	}
	if err := epc.LoadProposers(state); err != nil {
		return nil, err
	}
	return epc, nil
}

func (f *FinalizedChain) entryGetState(ctx context.Context, step Step) (*phase0.BeaconStateView, error) {
	f.RLock()
	defer f.RUnlock()
	return f.getState(ctx, step)
}

func (f *FinalizedChain) getState(ctx context.Context, step Step) (*phase0.BeaconStateView, error) {
	root := f.stateRoot(step)
	if root == (common.Root{}) {
		return nil, fmt.Errorf("unknown state, step out of range: %s", step)
	}
	state, exists, err := f.StateDB.Get(ctx, root)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("state for state-root %x (step %s) does not exist: %v", root, step, err)
	}
	return state, nil
}
