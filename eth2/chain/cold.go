package chain

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/db/states"
	"github.com/protolambda/ztyp/tree"
	"sync"
)

type ColdChain interface {
	Start() Slot
	End() Slot
	OnFinalizedEntry(entry *HotEntry) error
	Chain
}

type FinalizedEntryView struct {
	slot     Slot
	finChain *FinalizedChain
}

func (e *FinalizedEntryView) Slot() Slot {
	return e.slot
}

func (e *FinalizedEntryView) IsEmpty() bool {
	return e.ParentRoot() == e.BlockRoot()
}

func (e *FinalizedEntryView) ParentRoot() (root Root) {
	return e.finChain.parentRoot(e.slot)
}

func (e *FinalizedEntryView) BlockRoot() (root Root) {
	return e.finChain.blockRoot(e.slot)
}

func (e *FinalizedEntryView) StateRoot() Root {
	return e.finChain.stateRoot(e.slot)
}

func (e *FinalizedEntryView) EpochsContext(ctx context.Context) (*beacon.EpochsContext, error) {
	return e.finChain.getEpochsContext(ctx, e.slot)
}

func (e *FinalizedEntryView) State(ctx context.Context) (*beacon.BeaconStateView, error) {
	return e.finChain.getState(ctx, e.slot)
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
	PubkeyCache *beacon.PubkeyCache

	// Start of the historical data
	AnchorSlot Slot

	// Block roots, starting at AnchorSlot, indexed by slot.
	// A blockRoot can be a copy of the previous root if slot is empty
	BlockRoots []Root

	// StatePreBlockRoots are the state roots before applying a block, starting at AnchorSlot, indexed by slot.
	StatePreBlockRoots []Root

	// StatePostBlockRoots are the state roots after applying a block, starting at AnchorSlot, indexed by slot.
	// The root is the same as that in StatePreBlockRoots for gap slots.
	StatePostBlockRoots []Root

	// BlockRootsMap maps roots of BlockRoots back to slots
	BlockRootsMap map[Root]Slot

	// StatePreBlockRootsMap maps roots of StatePreBlockRoots back to slots
	StatePreBlockRootsMap map[Root]Slot

	// StatePostBlockRootsMap maps roots of StatePostBlockRoots back to slots
	StatePostBlockRootsMap map[Root]Slot

	// Spec is holds configuration information for the parameters and types of the chain
	Spec *beacon.Spec

	StateDB states.DB
}

var _ = ColdChain((*FinalizedChain)(nil))

func NewFinalizedChain(anchorSlot Slot, spec *beacon.Spec, stateDB states.DB) *FinalizedChain {
	initialCapacity := Slot(200)
	return &FinalizedChain{
		PubkeyCache:      beacon.EmptyPubkeyCache(),
		AnchorSlot:       anchorSlot,
		BlockRoots:       make([]Root, 0, initialCapacity),
		StatePreBlockRoots:       make([]Root, 0, initialCapacity),
		StatePostBlockRoots:       make([]Root, 0, initialCapacity),
		BlockRootsMap: make(map[Root]Slot, initialCapacity),
		StatePreBlockRootsMap: make(map[Root]Slot, initialCapacity),
		StatePostBlockRootsMap: make(map[Root]Slot, initialCapacity),
		Spec:             spec,
		StateDB:          stateDB,
	}
}

type ColdChainIter struct {
	Chain              Chain
	StartSlot, EndSlot Slot
	StartPreBlock, EndPreBlock bool
}

func (fi *ColdChainIter) Start() (slot Slot, preBlock bool) {
	return fi.StartSlot, fi.StartPreBlock
}

func (fi *ColdChainIter) End() (slot Slot, preBlock bool) {
	return fi.EndSlot, fi.EndPreBlock
}

func (fi *ColdChainIter) Entry(slot Slot, preBlock bool) (entry ChainEntry, err error) {
	start, startPreBl := fi.Start()
	end, endPreBl := fi.End()
	if slot < start || (slot == start && preBlock && !startPreBl) {
		return nil, fmt.Errorf("query too low")
	}
	if slot > end || (slot == end && !preBlock && endPreBl) {
		return nil, fmt.Errorf("query too high")
	}
	entry, err = fi.Chain.BySlot(slot, preBlock)
	return
}

func (f *FinalizedChain) Iter() (ChainIter, error) {
	start := f.Start()
	end := f.End()
	return &ColdChainIter{
		Chain:     f,
		StartSlot: start,
		EndSlot:   end,
	}, nil
}

// Start slot of the cold chain (inclusive)
func (f *FinalizedChain) Start() Slot {
	return f.AnchorSlot
}

// End slot of the cold chain part (exclusive), should equal the epoch start slot of the finalized checkpoint.
// The FinalizedChain is empty if Start() == End()
func (f *FinalizedChain) End() Slot {
	return f.AnchorSlot + Slot(len(f.BlockRoots))
}

var UnknownRootErr = errors.New("unknown root")

func (f *FinalizedChain) ByStateRoot(root Root) (entry ChainEntry, ok bool) {
	f.RLock()
	defer f.RUnlock()
	slot, ok := f.StatePreBlockRootsMap[root]
	if !ok {
		slot, ok = f.StatePostBlockRootsMap[root]
		if !ok {
			return nil, UnknownRootErr
		}
		f.BlockRoots
	}
	return f.bySlot(slot)
}

func (f *FinalizedChain) ByBlockRoot(root Root) (entry ChainEntry, ok bool) {
	f.RLock()
	defer f.RUnlock()
	return f.byBlockRoot(root)
}

func (f *FinalizedChain) byBlockRoot(root Root) (entry ChainEntry, ok bool) {
	slot, ok := f.SlotsByBlockRoot[root]
	if !ok {
		return nil, UnknownRootErr
	}
	return f.bySlot(slot)
}

func (f *FinalizedChain) Closest(fromBlockRoot Root, toSlot Slot) (ChainEntry, error) {
	f.RLock()
	defer f.RUnlock()
	if start := f.Start(); toSlot < start {
		return nil, fmt.Errorf("slot %d is too early. Start is at slot %d", toSlot, start)
	}
	// check if the root is canonical
	_, ok := f.SlotsByStateRoot[fromBlockRoot]
	if !ok {
		return nil, UnknownRootErr
	}
	// find the slot closest to the requested slot: whatever is still within range
	if end := f.End(); end == 0 {
		return nil, errors.New("empty chain, no data available")
	} else if toSlot >= end {
		toSlot = end - 1
	}
	return f.bySlot(toSlot)
}

func (f *FinalizedChain) IsAncestor(root Root, ofRoot Root) (unknown bool, isAncestor bool) {
	f.RLock()
	defer f.RUnlock()

	// can't be ancestors if they are equal.
	if root == ofRoot {
		return false, false
	}
	anchor, err := f.byBlockRoot(ofRoot)
	if err != nil {
		return true, false
	}
	lookup, err := f.byBlockRoot(root)
	if err != nil {
		return true, false
	}
	// if the nodes are the other way around,
	// they do not have the same ancestor relation, even though they are on the same chain.
	return true, anchor.Slot() < lookup.Slot()
}

func (f *FinalizedChain) BySlot(slot Slot, preBlock bool) (ChainEntry, error) {
	f.RLock()
	defer f.RUnlock()
	return f.bySlot(slot)
}

func (f *FinalizedChain) bySlot(slot Slot) (ChainEntry, error) {
	if start := f.Start(); slot < start {
		return nil, fmt.Errorf("slot %d is too early. Chain starts at slot %d", slot, start)
	}
	if end := f.End(); slot >= end {
		return nil, fmt.Errorf("slot %d is too late. Chain ends at slot %d", slot, end)
	}
	return &FinalizedEntryView{
		slot:     slot,
		finChain: f,
	}, nil
}

func (f *FinalizedChain) OnFinalizedEntry(entry *HotEntry) error {
	f.Lock()
	defer f.Unlock()
	if end := f.End(); entry.slot != end {
		return fmt.Errorf("expected next finalized entry to have slot %d, but got %d from entry with block root %s",
			end, entry.slot, entry.blockRoot.String())
	}
	postStateRoot := entry.state.HashTreeRoot(tree.GetHashFn())
	f.BlockRoots = append(f.BlockRoots, entry.blockRoot)
	f.StateRoots = append(f.StateRoots, postStateRoot)
	f.SlotsByStateRoot[postStateRoot] = entry.slot
	if entry.parentRoot != entry.blockRoot {
		// if it's not an empty slot, remember it by block root
		f.SlotsByBlockRoot[entry.blockRoot] = entry.slot
	}
	return nil
}

func (f *FinalizedChain) parentRoot(slot Slot) (root Root) {
	if slot <= f.AnchorSlot {
		return Root{}
	}
	return f.BlockRoots[slot-1-f.AnchorSlot]
}

func (f *FinalizedChain) blockRoot(slot Slot) (root Root) {
	if slot < f.AnchorSlot {
		return Root{}
	}
	i := slot - f.AnchorSlot
	if i >= Slot(len(f.BlockRoots)) {
		return Root{}
	}
	return f.BlockRoots[slot-f.AnchorSlot]
}

func (f *FinalizedChain) stateRoot(slot Slot) Root {
	if slot < f.AnchorSlot {
		return Root{}
	}
	i := slot - f.AnchorSlot
	if i >= Slot(len(f.StateRoots)) {
		return Root{}
	}
	return f.StateRoots[i]
}

func (f *FinalizedChain) getEpochsContext(ctx context.Context, slot Slot) (*beacon.EpochsContext, error) {
	epc := &beacon.EpochsContext{
		PubkeyCache: f.PubkeyCache,
	}
	// We do not store shuffling for older epochs
	// TODO: maybe store it after all, for archive node functionality?
	state, err := f.getState(ctx, slot)
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

func (f *FinalizedChain) getState(ctx context.Context, slot Slot) (*beacon.BeaconStateView, error) {
	root := f.stateRoot(slot)
	if root == (beacon.Root{}) {
		return nil, fmt.Errorf("unknown state, slot out of range: %d", slot)
	}
	state, exists, err := f.StateDB.Get(ctx, root)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("state for state-root %x (slot %d) does not exist: %v", root, slot, err)
	}
	return state, nil
}
