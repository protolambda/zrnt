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
	slot       Slot
	blockRoot  Root
	parentRoot Root
	epc        *beacon.EpochsContext
	state      *beacon.BeaconStateView
}

func NewHotEntry(
	slot Slot, blockRoot Root, parentRoot Root,
	state *beacon.BeaconStateView, epc *beacon.EpochsContext) *HotEntry {
	return &HotEntry{
		slot:       slot,
		epc:        epc,
		state:      state,
		blockRoot:  blockRoot,
		parentRoot: parentRoot,
	}
}

func (e *HotEntry) Slot() Slot {
	return e.slot
}

func (e *HotEntry) IsEmpty() bool {
	return e.parentRoot == Root{}
}

func (e *HotEntry) ParentRoot() (root Root) {
	return e.parentRoot
}

func (e *HotEntry) BlockRoot() (root Root) {
	return e.blockRoot
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

	// block++slot -> Entry
	Entries map[BlockSlotKey]*HotEntry
	// state root -> block+slot key
	State2Key map[Root]BlockSlotKey

	// BlockSink takes pruned entries and their canon status, and processes them.
	// Empty-slot entries will only occur for canonical chain,
	// non-canonical empty entries are ignored, as there can theoretically be an unlimited number of.
	// Non-canonical non-empty entries are still available, to track what is getting abandoned by the chain
	BlockSink BlockSink

	// Spec is holds configuration information for the parameters and types of the chain
	Spec *beacon.Spec
}

type HotChainIter struct {
	// ordered from head to 0
	entries  []*HotEntry
	headSlot Slot
}

func (fi *HotChainIter) Start() Slot {
	return fi.headSlot + 1 - Slot(len(fi.entries))
}

func (fi *HotChainIter) End() Slot {
	return fi.headSlot + 1
}

func (fi *HotChainIter) Entry(slot Slot) (entry ChainEntry, err error) {
	if slot < fi.Start() || slot >= fi.End() {
		return nil, fmt.Errorf("out of range slot: %d, range: [%d, %d)", slot, fi.Start(), fi.End())
	}
	return fi.entries[slot-fi.headSlot], nil
}

func (uc *UnfinalizedChain) Iter() (ChainIter, error) {
	headRef, err := uc.ForkChoice.FindHead()
	if err != nil {
		return nil, err
	}
	entries := make([]*HotEntry, 0)
	root := headRef.Root
	for {
		entry, err := uc.ByBlockRoot(root)
		if err != nil {
			break
		}
		entries = append(entries, entry.(*HotEntry))
		root = entry.ParentRoot()
	}
	return &HotChainIter{entries, headRef.Slot}, nil
}

type BlockSink interface {
	// Sink handles blocks that come from the Hot part, and may be finalized or not
	Sink(entry *HotEntry, canonical bool) error
}

type BlockSinkFn func(entry *HotEntry, canonical bool) error

func (fn BlockSinkFn) Sink(entry *HotEntry, canonical bool) error {
	return fn(entry, canonical)
}

func NewUnfinalizedChain(finalizedBlock *HotEntry, sink BlockSink, spec *beacon.Spec) (*UnfinalizedChain, error) {
	fin, err := finalizedBlock.state.FinalizedCheckpoint()
	if err != nil {
		return nil, err
	}
	finCh, err := fin.Raw()
	if err != nil {
		return nil, err
	}
	just, err := finalizedBlock.state.CurrentJustifiedCheckpoint()
	if err != nil {
		return nil, err
	}
	justCh, err := just.Raw()
	if err != nil {
		return nil, err
	}
	key := NewBlockSlotKey(finalizedBlock.blockRoot, finalizedBlock.slot)
	uc := &UnfinalizedChain{
		ForkChoice: nil,
		Entries:    map[BlockSlotKey]*HotEntry{key: finalizedBlock},
		State2Key:  map[Root]BlockSlotKey{finalizedBlock.StateRoot(): key},
		BlockSink:  sink,
		Spec:       spec,
	}
	uc.ForkChoice = forkchoice.NewForkChoice(
		finCh,
		justCh,
		forkchoice.BlockRef{Root: finalizedBlock.blockRoot, Slot: finalizedBlock.slot},
		finalizedBlock.parentRoot,
		forkchoice.BlockSinkFn(uc.OnPrunedBlock),
	)
	return uc, nil
}

func (uc *UnfinalizedChain) OnPrunedBlock(node *forkchoice.ProtoNode, canonical bool) error {
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

func (uc *UnfinalizedChain) ByStateRoot(root Root) (ChainEntry, error) {
	uc.RLock()
	defer uc.RUnlock()
	key, ok := uc.State2Key[root]
	if !ok {
		return nil, fmt.Errorf("unknown state %s", root)
	}
	return uc.byBlockSlot(key)
}

func (uc *UnfinalizedChain) ByBlockSlot(key BlockSlotKey) (ChainEntry, error) {
	uc.RLock()
	defer uc.RUnlock()
	return uc.byBlockSlot(key)
}

func (uc *UnfinalizedChain) byBlockSlot(key BlockSlotKey) (ChainEntry, error) {
	entry, ok := uc.Entries[key]
	if !ok {
		return nil, fmt.Errorf("unknown block slot, root: %s slot: %d", key.Root(), key.Slot())
	}
	return entry, nil
}

func (uc *UnfinalizedChain) ByBlockRoot(root Root) (ChainEntry, error) {
	uc.RLock()
	defer uc.RUnlock()
	ref, ok := uc.ForkChoice.GetBlock(root)
	if !ok {
		return nil, fmt.Errorf("unknown block %s", root)
	}
	return uc.byBlockSlot(NewBlockSlotKey(root, ref.Slot))
}

// Find closest block in subtree, up to given slot (may return entry of fromBlockRoot itself).
// Err if none, incl. fromBlockRoot, could be found.
func (uc *UnfinalizedChain) ClosestFrom(fromBlockRoot Root, toSlot Slot) (ChainEntry, error) {
	uc.RLock()
	defer uc.RUnlock()
	before, at, _, err := uc.ForkChoice.BlocksAroundSlot(fromBlockRoot, toSlot)
	if err != nil {
		return nil, err
	}
	if at != (forkchoice.BlockRef{}) {
		return uc.byBlockSlot(NewBlockSlotKey(at.Root, at.Slot))
	}
	if before != (forkchoice.BlockRef{}) {
		key := NewBlockSlotKey(before.Root, before.Slot)
		entry, ok := uc.Entries[key]
		if ok {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("could not find closest hot block starting from root %s, up to slot %d", fromBlockRoot, toSlot)
}

func (uc *UnfinalizedChain) BySlot(slot Slot) (ChainEntry, error) {
	uc.RLock()
	defer uc.RUnlock()
	_, at, _, err := uc.ForkChoice.BlocksAroundSlot(uc.Justified().Root, slot)
	if err != nil {
		return nil, err
	}
	if at == (forkchoice.BlockRef{}) {
		return nil, fmt.Errorf("no hot entry known for slot %d", slot)
	}
	if at.Slot != slot {
		panic("forkchoice bug, found block not actually at correct slot")
	}
	return uc.byBlockSlot(NewBlockSlotKey(at.Root, at.Slot))
}

func (uc *UnfinalizedChain) Justified() Checkpoint {
	return uc.ForkChoice.Justified()
}

func (uc *UnfinalizedChain) Finalized() Checkpoint {
	return uc.ForkChoice.Finalized()
}

func (uc *UnfinalizedChain) Head() (ChainEntry, error) {
	ref, err := uc.ForkChoice.FindHead()
	if err != nil {
		return nil, err
	}
	return uc.ByBlockRoot(ref.Root)
}

func (uc *UnfinalizedChain) AddBlock(ctx context.Context, signedBlock *beacon.SignedBeaconBlock) error {
	block := &signedBlock.Message
	blockRoot := block.HashTreeRoot(uc.Spec, tree.GetHashFn())

	pre, err := uc.ClosestFrom(block.ParentRoot, block.Slot)
	if err != nil {
		return err
	}

	if root := pre.BlockRoot(); root != block.ParentRoot {
		return fmt.Errorf("unknown parent root %s, found other root %s", block.ParentRoot, root)
	}

	epc, err := pre.EpochsContext(ctx)
	if err != nil {
		return err
	}

	state, err := pre.State(ctx)
	if err != nil {
		return err
	}

	// Process empty slots
	for slot := pre.Slot(); slot+1 < block.Slot; {
		if err := uc.Spec.ProcessSlot(ctx, state); err != nil {
			return err
		}
		// Per-epoch transition happens at the start of the first slot of every epoch.
		// (with the slot still at the end of the last epoch)
		isEpochEnd := uc.Spec.SlotToEpoch(slot+1) != uc.Spec.SlotToEpoch(slot)
		if isEpochEnd {
			if err := uc.Spec.ProcessEpoch(ctx, epc, state); err != nil {
				return err
			}
		}
		slot += 1
		if err := state.SetSlot(slot); err != nil {
			return err
		}
		if isEpochEnd {
			if err := epc.RotateEpochs(state); err != nil {
				return err
			}
		}

		// Add empty slot entry
		uc.Lock()
		uc.Entries[NewBlockSlotKey(block.ParentRoot, slot)] = &HotEntry{
			slot:       slot,
			epc:        epc,
			state:      state,
			blockRoot:  blockRoot,
			parentRoot: Root{},
		}
		uc.Unlock()

		state, err = beacon.AsBeaconStateView(state.Copy())
		if err != nil {
			return err
		}
		epc = epc.Clone()
	}

	// TODO: we're not storing the slot transition for the slot of the block itself (as if the slot was empty)
	if err := uc.Spec.StateTransition(ctx, epc, state, signedBlock, true); err != nil {
		return err
	}

	var finalizedEpoch, justifiedEpoch Epoch
	{
		finalizedCh, err := state.FinalizedCheckpoint()
		if err != nil {
			return err
		}
		finalizedEpoch, err = finalizedCh.Epoch()
		if err != nil {
			return err
		}
		justifiedCh, err := state.CurrentJustifiedCheckpoint()
		if err != nil {
			return err
		}
		justifiedEpoch, err = justifiedCh.Epoch()
		if err != nil {
			return err
		}
	}

	uc.Lock()
	defer uc.Unlock()
	uc.Entries[NewBlockSlotKey(blockRoot, block.Slot)] = &HotEntry{
		slot:       block.Slot,
		epc:        epc,
		state:      state,
		blockRoot:  blockRoot,
		parentRoot: block.ParentRoot,
	}
	uc.ForkChoice.ProcessBlock(
		forkchoice.BlockRef{Slot: block.Slot, Root: blockRoot},
		block.ParentRoot, justifiedEpoch, finalizedEpoch)

	if block.Slot < uc.AnchorSlot {
		uc.AnchorSlot = block.Slot
	}
	return nil
}

func (uc *UnfinalizedChain) AddAttestation(att *beacon.Attestation) error {
	blockRoot := att.Data.BeaconBlockRoot
	block, err := uc.ByBlockRoot(blockRoot)
	if err != nil {
		return err
	}
	_, ok := block.(*HotEntry)
	if !ok {
		return errors.New("expected HotEntry, need epochs-context to be present")
	}
	// HotEntry does not use a context, epochs-context is available.
	epc, err := block.EpochsContext(nil)
	if err != nil {
		return err
	}
	committee, err := epc.GetBeaconCommittee(att.Data.Slot, att.Data.Index)
	if err != nil {
		return err
	}
	indexedAtt, err := att.ConvertToIndexed(uc.Spec, committee)
	if err != nil {
		return err
	}
	targetEpoch := att.Data.Target.Epoch
	for _, index := range indexedAtt.AttestingIndices {
		uc.ForkChoice.ProcessAttestation(index, blockRoot, targetEpoch)
	}
	return nil
}
