package chain

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/forkchoice"
	"github.com/protolambda/zrnt/eth2/forkchoice/proto"
	"github.com/protolambda/ztyp/tree"
	"sync"
)

type HotEntry struct {
	self   BlockSlotKey
	parent Root
	epc    *common.EpochsContext
	state  common.BeaconState
}

func NewHotEntry(self BlockSlotKey, parent Root,
	state common.BeaconState, epc *common.EpochsContext) *HotEntry {
	return &HotEntry{
		self:   self,
		parent: parent,
		epc:    epc,
		state:  state,
	}
}

func (e *HotEntry) Step() Step {
	return AsStep(e.self.Slot, e.parent != e.self.Root)
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

func (e *HotEntry) EpochsContext(context.Context) (*common.EpochsContext, error) {
	return e.epc.Clone(), nil
}

func (e *HotEntry) State(context.Context) (common.BeaconState, error) {
	// Return a copy of the view, the state itself may not be modified
	return e.state.Copy()
}

type HotChain interface {
	Chain
	JustifiedCheckpoint() Checkpoint
	FinalizedCheckpoint() Checkpoint
	Justified() (ChainEntry, error)
	Finalized() (ChainEntry, error)
	Head() (ChainEntry, error)
	// First gets the closets ref from the given block root to the requested slot,
	// then transitions empty slots to get up to the requested slot.
	// A strict context should be provided to avoid costly long transitions.
	// An error is also returned if the fromBlockRoot is past the requested toSlot.
	Towards(ctx context.Context, fromBlockRoot Root, toSlot Slot) (ChainEntry, error)
	// Process a block. If there is an error, the chain is not mutated, and can be continued to use.
	AddBlock(ctx context.Context, signedBlock *phase0.SignedBeaconBlock) error
	// Process an attestation. If there is an error, the chain is not mutated, and can be continued to use.
	AddAttestation(att *phase0.Attestation) error
}

type UnfinalizedChain struct {
	sync.RWMutex

	ForkChoice forkchoice.Forkchoice

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
	Spec *common.Spec
}

var _ HotChain = (*UnfinalizedChain)(nil)

// Iterable over the unfinalized part of the chain (including the finalizing node at start of epoch).
// The view stays consistent during iteration: it's a full shallow copy of the canonical branch of the tree.
// Each slot is represented with two consecutive entry pointers.
// First for the preBlock node (omitted if not part of the chain at the start),
// second for the postBlock node (and is nil if it's an empty slot, omitted if not part of the chain at the end).
type HotChainIter []*HotEntry // Ordered from finalized slot to head slot

var _ ChainIter = (HotChainIter)(nil)

func (fi HotChainIter) Start() Step {
	return fi[0].Step()
}

func (fi HotChainIter) End() Step {
	return fi[len(fi)-1].Step()
}

func (fi HotChainIter) Entry(step Step) (entry ChainEntry, err error) {
	start := fi.Start()
	if step < start {
		return nil, fmt.Errorf("query too low")
	}
	end := fi.End()
	if step >= end {
		return nil, fmt.Errorf("query too high")
	}
	i := step - start
	return fi[i], nil
}

func (uc *UnfinalizedChain) Iter() (ChainIter, error) {
	uc.Lock()
	defer uc.Unlock()
	fin := uc.ForkChoice.Finalized()
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
		// if this is the first node of the (pre, post) pair, then check for the 2nd, and add nil if there is none.
		if node.Root == node.ParentRoot {
			if i+1 < len(nodes) { // clip at the end.
				next := nodes[i+1]
				if next.Slot != node.Slot {
					entries = append(entries, nil)
				}
			}
		}
	}

	return HotChainIter(entries), nil
}

type BlockSink interface {
	// Sink handles blocks that come from the Hot part, and may be finalized or not
	Sink(ctx context.Context, entry ChainEntry, canonical bool) error
}

type BlockSinkFn func(ctx context.Context, entry ChainEntry, canonical bool) error

func (fn BlockSinkFn) Sink(ctx context.Context, entry ChainEntry, canonical bool) error {
	return fn(ctx, entry, canonical)
}

func NewUnfinalizedChain(anchorState *phase0.BeaconStateView, sink BlockSink, spec *common.Spec) (*UnfinalizedChain, error) {
	fin, err := anchorState.FinalizedCheckpoint()
	if err != nil {
		return nil, err
	}
	just, err := anchorState.CurrentJustifiedCheckpoint()
	if err != nil {
		return nil, err
	}

	latestHeader, err := anchorState.LatestBlockHeader()
	if err != nil {
		return nil, err
	}
	if latestHeader.StateRoot == (Root{}) {
		latestHeader.StateRoot = anchorState.HashTreeRoot(tree.GetHashFn())
	}
	anchorBlockRoot := latestHeader.HashTreeRoot(tree.GetHashFn())

	slot, err := anchorState.Slot()
	if err != nil {
		return nil, err
	}
	epc, err := common.NewEpochsContext(spec, anchorState)
	if err != nil {
		return nil, err
	}
	anchor := BlockSlotKey{Root: anchorBlockRoot, Slot: slot}
	anchorBlock := &HotEntry{
		self: anchor,
		// parent root may be equal to anchorBlockRoot if the anchor state is of a gap slot.
		parent: latestHeader.ParentRoot,
		epc:    epc,
		state:  anchorState,
	}
	uc := &UnfinalizedChain{
		ForkChoice: nil,
		Entries:    map[BlockSlotKey]*HotEntry{anchor: anchorBlock},
		State2Key:  map[Root]BlockSlotKey{latestHeader.StateRoot: anchor},
		BlockSink:  sink,
		Spec:       spec,
	}
	balancesView, err := anchorState.Balances()
	if err != nil {
		return nil, err
	}
	balances, err := balancesView.AllBalances()
	if err != nil {
		return nil, err
	}
	fc, err := proto.NewProtoForkChoice(
		spec,
		fin,
		just,
		anchorBlockRoot, slot,
		latestHeader.ParentRoot,
		balances,
		proto.NodeSinkFn(uc.onPrunedNode),
	)
	if err != nil {
		return nil, err
	}
	uc.ForkChoice = fc
	return uc, nil
}

// onPrunedNode handles when nodes leave the forkchoice, and thus get removed from the hot view of the chain.
// Includes empty slots and nodes of the slot pre-block processing (even if the block exists)
func (uc *UnfinalizedChain) onPrunedNode(ctx context.Context, ref forkchoice.NodeRef, canonical bool) error {
	// Does not lock the hot chain again: the only caller is the forkchoice, internal to this hot chain,
	// which is always locked when the forkchoice pruning runs.
	key := BlockSlotKey{Slot: ref.Slot, Root: ref.Root}
	entry, ok := uc.Entries[key]
	if !ok {
		return nil
	}
	// Remove node from hot state
	delete(uc.Entries, key)
	delete(uc.State2Key, entry.StateRoot())
	// Move the node to the sink.
	return uc.BlockSink.Sink(ctx, entry, canonical)
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

func (uc *UnfinalizedChain) ByBlock(root Root) (entry ChainEntry, ok bool) {
	uc.RLock()
	defer uc.RUnlock()
	slot, ok := uc.ForkChoice.GetSlot(root)
	if !ok {
		return nil, false
	}
	return uc.byBlockSlot(BlockSlotKey{Slot: slot, Root: root})
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

func (uc *UnfinalizedChain) Search(parentRoot *Root, slot *Slot) ([]SearchEntry, error) {
	uc.Lock()
	defer uc.Unlock()
	finalized := uc.ForkChoice.Finalized()
	finSlot, _ := uc.Spec.EpochStartSlot(finalized.Epoch)
	anchor := common.NodeRef{Root: finalized.Root, Slot: finSlot}
	nonCanon, canon, err := uc.ForkChoice.Search(anchor, parentRoot, slot)
	if err != nil {
		return nil, err
	}
	out := make([]SearchEntry, 0, len(nonCanon)+len(canon))
	for _, ref := range canon {
		entry, ok := uc.byBlockSlot(BlockSlotKey{Root: ref.Root, Slot: ref.Slot})
		if ok {
			out = append(out, SearchEntry{ChainEntry: entry, Canonical: true})
		}
	}
	for _, ref := range nonCanon {
		entry, ok := uc.byBlockSlot(BlockSlotKey{Root: ref.Root, Slot: ref.Slot})
		if ok {
			out = append(out, SearchEntry{ChainEntry: entry, Canonical: false})
		}
	}
	return out, nil
}

func (uc *UnfinalizedChain) Closest(fromBlockRoot Root, toSlot Slot) (entry ChainEntry, ok bool) {
	uc.RLock()
	defer uc.RUnlock()
	return uc.closest(fromBlockRoot, toSlot)
}

func (uc *UnfinalizedChain) closest(fromBlockRoot Root, toSlot Slot) (entry ChainEntry, ok bool) {
	ref, err := uc.ForkChoice.ClosestToSlot(fromBlockRoot, toSlot)
	if err != nil {
		return nil, false
	}
	return uc.byBlockSlot(BlockSlotKey{Root: ref.Root, Slot: ref.Slot})
}

// helper function to fetch justified and finalized checkpoint from a beacon state
func stateJustFin(state *phase0.BeaconStateView) (justified Checkpoint, finalized Checkpoint, err error) {
	justifiedCh, err := state.CurrentJustifiedCheckpoint()
	if err != nil {
		return Checkpoint{}, Checkpoint{}, err
	}
	finalizedCh, err := state.FinalizedCheckpoint()
	if err != nil {
		return Checkpoint{}, Checkpoint{}, err
	}
	return justifiedCh, finalizedCh, nil
}

func (uc *UnfinalizedChain) Towards(ctx context.Context, fromBlockRoot Root, toSlot Slot) (ChainEntry, error) {
	uc.Lock()
	defer uc.Unlock()
	closest, ok := uc.closest(fromBlockRoot, toSlot)
	if !ok {
		return nil, fmt.Errorf("failed to find starting point to root %s to go towards slot %d", fromBlockRoot, toSlot)
	}
	if closest.Step().Slot() == toSlot {
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
	for slot := closest.Step().Slot(); slot < toSlot; {
		if err := phase0.ProcessSlot(ctx, uc.Spec, state); err != nil {
			return nil, err
		}
		// Per-epoch transition happens at the start of the first slot of every epoch.
		// (with the slot still at the end of the last epoch)
		isEpochEnd := uc.Spec.SlotToEpoch(slot+1) != uc.Spec.SlotToEpoch(slot)
		if isEpochEnd {
			if err := phase0.ProcessEpoch(ctx, uc.Spec, epc, state); err != nil {
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
		justified, finalized, err := stateJustFin(state)
		if err != nil {
			return nil, err
		}
		// Make the forkchoice aware of this new slot
		uc.ForkChoice.ProcessSlot(fromBlockRoot, slot, justified.Epoch, finalized.Epoch)
		// Make the forkchoice aware of latest justified/finalized data. Lazy-fetch the balances if necessary.
		if err := uc.ForkChoice.UpdateJustified(ctx, fromBlockRoot, finalized, justified,
			func() ([]forkchoice.Gwei, error) {
				balancesView, err := state.Balances()
				if err != nil {
					return nil, err
				}
				return balancesView.AllBalances()
			}); err != nil {
			return nil, fmt.Errorf("failed to update forkchoice with new justification data: %v", err)
		}

		// Track the entry
		key := BlockSlotKey{Root: fromBlockRoot, Slot: slot}
		entry := &HotEntry{
			self:   key,
			epc:    epc,
			state:  state,
			parent: fromBlockRoot,
		}
		uc.Entries[key] = entry
		stateRoot := state.HashTreeRoot(tree.GetHashFn())
		uc.State2Key[stateRoot] = key
		last = entry

		state, err = phase0.AsBeaconStateView(state.Copy())
		if err != nil {
			return nil, err
		}
		epc = epc.Clone()
	}
	return last, nil
}

func (uc *UnfinalizedChain) InSubtree(anchor Root, root Root) (unknown bool, inSubtree bool) {
	uc.RLock()
	defer uc.RUnlock()
	return uc.ForkChoice.InSubtree(anchor, root)
}

func (uc *UnfinalizedChain) ByCanonStep(step Step) (entry ChainEntry, ok bool) {
	uc.Lock()
	defer uc.Unlock()
	ref, err := uc.ForkChoice.CanonAtSlot(uc.ForkChoice.Justified().Root, step.Slot(), step.Block())
	if err != nil {
		return nil, false
	}
	return uc.byBlockSlot(BlockSlotKey{Root: ref.Root, Slot: ref.Slot})
}

func (uc *UnfinalizedChain) JustifiedCheckpoint() Checkpoint {
	uc.RLock()
	defer uc.RUnlock()
	return uc.ForkChoice.Justified()
}

func (uc *UnfinalizedChain) FinalizedCheckpoint() Checkpoint {
	uc.RLock()
	defer uc.RUnlock()
	return uc.ForkChoice.Finalized()
}

func (uc *UnfinalizedChain) Justified() (ChainEntry, error) {
	uc.RLock()
	defer uc.RUnlock()
	justified := uc.ForkChoice.Justified()
	slot, _ := uc.Spec.EpochStartSlot(justified.Epoch)
	entry, ok := uc.byBlockSlot(BlockSlotKey{Root: justified.Root, Slot: slot})
	if !ok {
		return nil, fmt.Errorf("forkchoice found justified node that is not in the hot chain: %s:%d",
			justified.Root, slot)
	}
	return entry, nil
}

func (uc *UnfinalizedChain) Finalized() (ChainEntry, error) {
	uc.RLock()
	defer uc.RUnlock()
	finalized := uc.ForkChoice.Finalized()
	slot, _ := uc.Spec.EpochStartSlot(finalized.Epoch)
	entry, ok := uc.byBlockSlot(BlockSlotKey{Root: finalized.Root, Slot: slot})
	if !ok {
		return nil, fmt.Errorf("forkchoice found finalized node that is not in the hot chain: %s:%d",
			finalized.Root, slot)
	}
	return entry, nil
}

func (uc *UnfinalizedChain) Head() (ChainEntry, error) {
	uc.Lock()
	defer uc.Unlock()
	ref, err := uc.ForkChoice.Head()
	if err != nil {
		return nil, err
	}
	entry, ok := uc.byBlockSlot(BlockSlotKey{Root: ref.Root, Slot: ref.Slot})
	if !ok {
		return nil, fmt.Errorf("forkchoice found head node that is not in the hot chain: %s:%d",
			ref.Root, ref.Slot)
	}
	return entry, nil
}

func (uc *UnfinalizedChain) AddBlock(ctx context.Context, signedBlock *phase0.SignedBeaconBlock) error {
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
	if err := beacon.PostSlotTransition(ctx, uc.Spec, epc, state, signedBlock, true); err != nil {
		return err
	}

	justified, finalized, err := stateJustFin(state)
	if err != nil {
		return err
	}

	// Make the forkchoice aware of the new block
	uc.ForkChoice.ProcessBlock(block.ParentRoot, blockRoot, block.Slot, justified.Epoch, finalized.Epoch)

	key := BlockSlotKey{Slot: block.Slot, Root: blockRoot}
	uc.Entries[key] = &HotEntry{
		self:   key,
		parent: block.ParentRoot,
		epc:    epc,
		state:  state,
	}
	uc.State2Key[block.StateRoot] = key

	return nil
}

// AddAttestation updates the forkchoice with the given attestation.
// Warning: the attestation signature is not verified, it is up to the caller to verify.
func (uc *UnfinalizedChain) AddAttestation(att *phase0.Attestation) error {
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
		// the node should exist, unless pruned during attestation processing, fine to ignore.
		_ = uc.ForkChoice.ProcessAttestation(index, data.BeaconBlockRoot, data.Slot)
	}
	return nil
}
