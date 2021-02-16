package chain

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/forkchoice"
	"github.com/protolambda/ztyp/tree"
	"sort"
	"sync"
)

type HotEntry struct {
	self   BlockSlotKey
	parent Root
	epc    *beacon.EpochsContext
	state  *beacon.BeaconStateView
}

func NewHotEntry(self BlockSlotKey, parent Root,
	state *beacon.BeaconStateView, epc *beacon.EpochsContext) *HotEntry {
	return &HotEntry{
		self:   self,
		parent: parent,
		epc:    epc,
		state:  state,
	}
}

func (e *HotEntry) Slot() Slot {
	return e.self.Slot
}

func (e *HotEntry) IsEmpty() bool {
	return e.parent == e.self.Root
}

func (e *HotEntry) ParentRoot() (root Root) {
	return e.parent
}

func (e *HotEntry) BlockRoot() (root Root) {
	return e.self.Root
}

func (e *HotEntry) StateRoot() Root {
	return e.state.HashTreeRoot(tree.GetHashFn())
}

func (e *HotEntry) EpochsContext(ctx context.Context) (*beacon.EpochsContext, error) {
	return e.epc.Clone(), nil
}

func (e *HotEntry) State(ctx context.Context) (*beacon.BeaconStateView, error) {
	// Return a copy of the view, the state itself may not be modified
	return beacon.AsBeaconStateView(e.state.Copy())
}

type HotChain interface {
	Chain
	Justified() Checkpoint
	Finalized() Checkpoint
	Head() (ChainEntry, error)
	// Process a block. If there is an error, the chain is not mutated, and can be continued to use.
	AddBlock(ctx context.Context, signedBlock *beacon.SignedBeaconBlock) error
	// Process an attestation. If there is an error, the chain is not mutated, and can be continued to use.
	AddAttestation(att *beacon.Attestation) error
}

type UnfinalizedChain struct {
	sync.RWMutex

	ForkChoice *forkchoice.ForkChoice

	AnchorSlot Slot

	// Block root (parent if empty slot) and slot -> Entry
	Entries map[BlockSlotKey]*HotEntry

	// State root -> block+slot key
	//
	// State roots here include the updated latest-header, and matches the state root in the block.
	// For empty slots, they match the state root after slot processing.
	State2Key map[Root]BlockSlotKey

	// BlockSink takes pruned entries and their canon status, and processes them.
	// Empty-slot entries will only occur for canonical chain,
	// non-canonical empty entries are ignored, as there can theoretically be an unlimited number of.
	// Non-canonical non-empty entries are still available, to track what is getting abandoned by the chain
	BlockSink BlockSink

	// Spec is holds configuration information for the parameters and types of the chain
	Spec *beacon.Spec
}

// ordered from finalized slot to head slot
type HotChainIter []*HotEntry

func (fi HotChainIter) Start() Slot {
	return fi[0].slot
}

func (fi HotChainIter) End() Slot {
	return fi[len(fi)-1].slot
}

func (fi HotChainIter) Entry(slot Slot) (entry ChainEntry, err error) {
	start, end := fi.Start(), fi.End()
	if slot < start || slot >= end {
		return nil, fmt.Errorf("out of range slot: %d, range: [%d, %d)", slot, fi.Start(), fi.End())
	}
	return fi[slot-start], nil
}

func (uc *UnfinalizedChain) Iter() (ChainIter, error) {
	uc.Lock()
	defer uc.Unlock()
	fin := uc.Finalized()
	finSlot, _ := uc.Spec.EpochStartSlot(fin.Epoch)
	// block nodes also have gap slots. Reduce that back to normal.
	nodes, err := uc.ForkChoice.CanonicalChain(fin.Root, finSlot)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, errors.New("empty chain")
	}
	entries := make([]*HotEntry, 0, len(nodes))
	for i := 0; i < len(nodes); i++ {
		node := &nodes[i]
		entry, ok := uc.Entries[BlockSlotKey{Root: node.Root, Slot: node.Slot}]
		if !ok {
			return nil, fmt.Errorf("missing hot entry for node root %s slot %d", node.Root, node.Slot)
		}
		entries = append(entries, entry)
	}

	return HotChainIter(entries), nil
}

type BlockSink interface {
	// Sink handles blocks that come from the Hot part, and may be finalized or not
	Sink(entry *HotEntry, canonical bool) error
}

type BlockSinkFn func(entry *HotEntry, canonical bool) error

func (fn BlockSinkFn) Sink(entry *HotEntry, canonical bool) error {
	return fn(entry, canonical)
}

func NewUnfinalizedChain(anchorState *beacon.BeaconStateView, sink BlockSink, spec *beacon.Spec) (*UnfinalizedChain, error) {
	fin, err := anchorState.FinalizedCheckpoint()
	if err != nil {
		return nil, err
	}
	finCh, err := fin.Raw()
	if err != nil {
		return nil, err
	}
	just, err := anchorState.CurrentJustifiedCheckpoint()
	if err != nil {
		return nil, err
	}
	justCh, err := just.Raw()
	if err != nil {
		return nil, err
	}

	latestHeader, err := anchorState.LatestBlockHeader()
	if err != nil {
		return nil, err
	}
	latestHeader, _ = beacon.AsBeaconBlockHeader(latestHeader.Copy())
	stateRoot, err := latestHeader.StateRoot()
	if err != nil {
		return nil, err
	}
	if stateRoot == (Root{}) {
		stateRoot = anchorState.HashTreeRoot(tree.GetHashFn())
		if err := latestHeader.SetStateRoot(stateRoot); err != nil {
			return nil, err
		}
	}
	anchorBlockRoot := latestHeader.HashTreeRoot(tree.GetHashFn())

	slot, err := anchorState.Slot()
	if err != nil {
		return nil, err
	}
	// may be equal to anchorBlockRoot if the anchor state is of a gap slot.
	parentRoot, err := latestHeader.ParentRoot()
	if err != nil {
		return nil, err
	}
	epc, err := spec.NewEpochsContext(anchorState)
	if err != nil {
		return nil, err
	}
	anchor := BlockSlotKey{Root: anchorBlockRoot, Slot: slot}
	anchorBlock := &HotEntry{
		self:   anchor,
		parent: parentRoot,
		epc:    epc,
		state:  anchorState,
	}
	uc := &UnfinalizedChain{
		ForkChoice: nil,
		Entries:    map[BlockSlotKey]*HotEntry{anchor: anchorBlock},
		State2Key:  map[Root]BlockSlotKey{stateRoot: anchor},
		BlockSink:  sink,
		Spec:       spec,
	}
	uc.ForkChoice = forkchoice.NewForkChoice(
		spec,
		finCh,
		justCh,
		anchorBlockRoot, slot,
		parentRoot,
		forkchoice.BlockSinkFn(uc.OnPrunedNode),
	)
	return uc, nil
}

// TODO
func (uc *UnfinalizedChain) OnPrunedNode(node *forkchoice.ProtoNode, canonical bool) error {
	uc.Lock()
	defer uc.Unlock()
	blockRef := node.Block

	key := NewBlockSlotKey(blockRef.Root, blockRef.Slot)
	entry, ok := uc.Entries[key]
	if ok {
		// Remove block from hot state
		delete(uc.Entries, key)
		delete(uc.State2Key, entry.StateRoot())
		// There may be empty slots leading up to the block,
		// If this block is not canonical, we cannot delete them,
		// because a later block may still share the history, and be canonical.
		// So we only prune if we find canonical blocks that get pruned.
		if canonical {
			// If pruning this entry means we prune something after the anchor,
			// adjust the anchor to the first slot after what was pruned.
			if entry.slot+1 > uc.AnchorSlot {
				uc.AnchorSlot = entry.slot + 1
			}
			// remove every entry before this pruned block
			pruned := make([]*HotEntry, 0)
			for _, e := range uc.Entries {
				// TODO: more aggressive pruning, we don't need every branch. Filter everything not in correct subtree.
				if e.slot < uc.AnchorSlot {
					pruned = append(pruned, e)
				}
			}
			// sink from oldest to newest entry
			sort.Slice(pruned, func(i, j int) bool {
				return pruned[i].slot < pruned[j].slot
			})
			for _, e := range pruned {
				delete(uc.Entries, NewBlockSlotKey(e.blockRoot, e.slot))
				delete(uc.State2Key, e.StateRoot())
				if err := uc.BlockSink.Sink(e, true); err != nil {
					return err
				}
			}
		} else {
			// Only sink the actual block that was pruned, if non-canonical.
			if err := uc.BlockSink.Sink(entry, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (uc *UnfinalizedChain) ByStateRoot(root Root) (entry ChainEntry, ok bool) {
	uc.RLock()
	defer uc.RUnlock()
	key, ok := uc.State2Key[root]
	if !ok {
		return nil, false
	}
	return uc.byBlockSlot(key)
}

func (uc *UnfinalizedChain) byBlockSlot(key BlockSlotKey) (entry ChainEntry, ok bool) {
	entry, ok = uc.Entries[key]
	return entry, ok
}

func (uc *UnfinalizedChain) ByBlockSlot(root Root, slot Slot) (entry ChainEntry, ok bool) {
	uc.RLock()
	defer uc.RUnlock()
	return uc.byBlockSlot(BlockSlotKey{Slot: slot, Root: root})
}

func (uc *UnfinalizedChain) Closest(fromBlockRoot Root, toSlot Slot) (ChainEntry, error) {
	uc.RLock()
	defer uc.RUnlock()
	return uc.closest(fromBlockRoot, toSlot)
}

func (uc *UnfinalizedChain) closest(fromBlockRoot Root, toSlot Slot) (ChainEntry, error) {
	ref, err := uc.ForkChoice.ClosestToSlot(fromBlockRoot, toSlot)
	if err != nil {
		return nil, err
	}
	if ref != (forkchoice.NodeRef{}) {
		entry, ok := uc.byBlockSlot(BlockSlotKey{Root: ref.Root, Slot: ref.Slot})
		if !ok {
			panic("node disappeared (unreachable)")
		}
		return entry, nil
	}
	return nil, fmt.Errorf("could not find closest hot block starting from root %s, up to slot %d", fromBlockRoot, toSlot)
}

// helper function to fetch justified and finalized epoch from a beacon state
func stateJustFin(state *beacon.BeaconStateView) (justifiedEpoch Epoch, finalizedEpoch Epoch, err error) {
	justifiedCh, err := state.CurrentJustifiedCheckpoint()
	if err != nil {
		return 0, 0, err
	}
	justifiedEpoch, err = justifiedCh.Epoch()
	if err != nil {
		return 0, 0, err
	}
	finalizedCh, err := state.FinalizedCheckpoint()
	if err != nil {
		return 0, 0, err
	}
	finalizedEpoch, err = finalizedCh.Epoch()
	if err != nil {
		return 0, 0, err
	}
	return
}

func (uc *UnfinalizedChain) Towards(ctx context.Context, fromBlockRoot Root, toSlot Slot) (ChainEntry, error) {
	uc.Lock()
	defer uc.Unlock()
	closest, err := uc.closest(fromBlockRoot, toSlot)
	if err != nil {
		return nil, err
	}
	if closest.Slot() == toSlot {
		return closest, nil
	}

	epc, err := closest.EpochsContext(ctx)
	if err != nil {
		return nil, err
	}

	state, err := closest.State(ctx)
	if err != nil {
		return nil, err
	}

	var last *HotEntry
	// Process empty slots
	for slot := closest.Slot(); slot < toSlot; {
		if err := uc.Spec.ProcessSlot(ctx, state); err != nil {
			return nil, err
		}
		// Per-epoch transition happens at the start of the first slot of every epoch.
		// (with the slot still at the end of the last epoch)
		isEpochEnd := uc.Spec.SlotToEpoch(slot+1) != uc.Spec.SlotToEpoch(slot)
		if isEpochEnd {
			if err := uc.Spec.ProcessEpoch(ctx, epc, state); err != nil {
				return nil, err
			}
		}
		slot += 1
		if err := state.SetSlot(slot); err != nil {
			return nil, err
		}
		if isEpochEnd {
			if err := epc.RotateEpochs(state); err != nil {
				return nil, err
			}
		}
		justifiedEpoch, finalizedEpoch, err := stateJustFin(state)
		if err != nil {
			return nil, err
		}
		// Make the forkchoice aware of this new slot
		uc.ForkChoice.ProcessSlot(fromBlockRoot, slot, justifiedEpoch, finalizedEpoch)

		// Track the entry
		key := BlockSlotKey{Root: fromBlockRoot, Slot: slot}
		entry := &HotEntry{
			self:   key,
			epc:    epc,
			state:  state,
			parent: fromBlockRoot,
		}
		uc.Entries[key] = entry
		last = entry

		state, err = beacon.AsBeaconStateView(state.Copy())
		if err != nil {
			return nil, err
		}
		epc = epc.Clone()
	}
	return last, nil
}

func (uc *UnfinalizedChain) IsAncestor(root Root, ofRoot Root) (unknown bool, isAncestor bool) {
	uc.RLock()
	defer uc.RUnlock()
	return uc.ForkChoice.IsAncestor(root, ofRoot)
}

func (uc *UnfinalizedChain) BySlot(slot Slot) (ChainEntry, error) {
	uc.Lock()
	defer uc.Unlock()
	closest, err := uc.ForkChoice.CanonAtSlot(uc.Justified().Root, slot)
	if err != nil {
		return nil, err
	}
	if closest.Slot != slot {
		return nil, fmt.Errorf("cannot find node at the given slot %d, but found to %d", slot, closest.Slot)
	}
	entry, ok := uc.byBlockSlot(BlockSlotKey{Root: closest.Root, Slot: slot})
	if !ok {
		return nil, fmt.Errorf("forkchoice found node not present in hot chain: %s:%d", closest.Root, closest.Slot)
	}
	return entry, nil
}

func (uc *UnfinalizedChain) Justified() Checkpoint {
	return uc.ForkChoice.Justified()
}

func (uc *UnfinalizedChain) Finalized() Checkpoint {
	return uc.ForkChoice.Finalized()
}

func (uc *UnfinalizedChain) Head() (ChainEntry, error) {
	uc.Lock()
	defer uc.Unlock()
	ref, err := uc.ForkChoice.FindHead()
	if err != nil {
		return nil, err
	}
	entry, ok := uc.byBlockSlot(BlockSlotKey{Root: ref.Root, Slot: ref.Slot})
	if !ok {
		return nil, fmt.Errorf("forkchoice found head node not in hot chain: %s:%d", ref.Root, ref.Slot)
	}
	return entry, nil
}

func (uc *UnfinalizedChain) AddBlock(ctx context.Context, signedBlock *beacon.SignedBeaconBlock) error {
	uc.Lock()
	defer uc.Unlock()

	block := &signedBlock.Message
	pre, err := uc.Towards(ctx, block.ParentRoot, block.Slot)
	if err != nil {
		return fmt.Errorf("failed to prepare for block, towards-slot failed: %v", err)
	}

	blockRoot := block.HashTreeRoot(uc.Spec, tree.GetHashFn())

	state, err := pre.State(ctx)
	if err != nil {
		return err
	}
	epc, err := pre.EpochsContext(ctx)
	if err != nil {
		return err
	}

	// we already processed the slots (including that of the block itself), just finish the transition.
	if err := uc.Spec.PostSlotTransition(ctx, epc, state, signedBlock, true); err != nil {
		return err
	}

	justifiedEpoch, finalizedEpoch, err := stateJustFin(state)
	if err != nil {
		return err
	}

	// Make the forkchoice aware of the new block
	uc.ForkChoice.ProcessBlock(block.ParentRoot, blockRoot, block.Slot, justifiedEpoch, finalizedEpoch)

	key := BlockSlotKey{Slot: block.Slot, Root: blockRoot}
	uc.Entries[key] = &HotEntry{
		self:   key,
		parent: block.ParentRoot,
		epc:    epc,
		state:  state,
	}

	return nil
}

func (uc *UnfinalizedChain) AddAttestation(att *beacon.Attestation) error {
	uc.Lock()
	defer uc.Unlock()

	data := &att.Data
	node, ok := uc.ByBlockSlot(data.BeaconBlockRoot, data.Slot)
	if ok {
		return fmt.Errorf("unknown block and slot pair: %s, %d", data.BeaconBlockRoot, data.Slot)
	}
	epc, err := node.EpochsContext(context.Background())
	if err != nil {
		return err
	}
	committee, err := epc.GetBeaconCommittee(data.Slot, data.Index)
	if err != nil {
		return err
	}
	indexedAtt, err := att.ConvertToIndexed(uc.Spec, committee)
	if err != nil {
		return err
	}
	for _, index := range indexedAtt.AttestingIndices {
		uc.ForkChoice.ProcessAttestation(index, data.BeaconBlockRoot, data.Slot)
	}
	return nil
}
