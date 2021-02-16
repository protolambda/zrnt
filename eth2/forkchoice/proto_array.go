package forkchoice

import (
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"sync"
)

type Root = beacon.Root
type Epoch = beacon.Epoch
type Slot = beacon.Slot
type ValidatorIndex = beacon.ValidatorIndex
type Gwei = beacon.Gwei
type Checkpoint = beacon.Checkpoint

type SignedGwei = int64

type ExtendedRef struct {
	Slot Slot
	// Block root, may be equal to parent root if empty
	Root       Root
	ParentRoot Root
}

type NodeRef struct {
	Slot Slot
	// Block root, may be equal to parent root if empty
	Root Root
}

type ProtoNodeIndex uint64

const NONE = ^ProtoNodeIndex(0)

type ProtoNode struct {
	Ref    NodeRef
	Parent ProtoNodeIndex
	// Duplicated to avoid pruning of this useful info.
	ParentRoot     Root
	JustifiedEpoch Epoch
	FinalizedEpoch Epoch
	Weight         SignedGwei
	BestChild      ProtoNodeIndex
	BestDescendant ProtoNodeIndex
}

type BlockSinkFn func(ref NodeRef, canonical bool) error

func (fn BlockSinkFn) OnPrunedNode(ref NodeRef, canonical bool) error {
	return fn(ref, canonical)
}

type NodeSink interface {
	OnPrunedNode(ref NodeRef, canonical bool) error
}

// Tracks slots and blocks as nodes.
// Every block has two nodes: with and without the block. The node with the block is the child of that without it.
// Gap slots just have a single node.
// There may be multiple nodes with the same parent but different blocks (i.e. double proposals, but slashable).
type ProtoArray struct {
	sink           NodeSink
	indexOffset    ProtoNodeIndex
	justifiedEpoch Epoch
	finalizedEpoch Epoch
	nodes          []ProtoNode
	// maintains only nodes that are actually part of the tree starting from finalized point.
	indices map[NodeRef]ProtoNodeIndex
	// Tracks the first slot at or after the block root that the array knows of.
	// The lowest slot for a block does not equal the block.slot itself, that may have been pruned.
	blockSlots         map[Root]Slot
	updatedConnections bool
}

func NewProtoArray(justifiedEpoch Epoch, finalizedEpoch Epoch, sink NodeSink) *ProtoArray {
	arr := ProtoArray{
		sink:               sink,
		indexOffset:        0,
		justifiedEpoch:     justifiedEpoch,
		finalizedEpoch:     finalizedEpoch,
		nodes:              make([]ProtoNode, 0, 100),
		indices:            make(map[NodeRef]ProtoNodeIndex, 100),
		updatedConnections: true,
	}
	return &arr
}

var invalidIndexErr = errors.New("invalid index")

func (pr *ProtoArray) getNode(index ProtoNodeIndex) (*ProtoNode, error) {
	if index < pr.indexOffset {
		return nil, invalidIndexErr
	}
	i := index - pr.indexOffset
	if i > ProtoNodeIndex(len(pr.nodes)) {
		return nil, invalidIndexErr
	}
	return &pr.nodes[i], nil
}

// From head back to anchor root (including the anchor itself, if present) and anchor slot.
// Includes nodes with empty block, then followed up by a node with the block if there is any.
func (pr *ProtoArray) CanonicalChain(anchorRoot Root, anchorSlot Slot) ([]ExtendedRef, error) {
	head, err := pr.FindHead(anchorRoot, anchorSlot)
	if err != nil {
		return nil, err
	}
	chain := make([]ExtendedRef, 0, len(pr.nodes))
	index := pr.indices[head]
	for index != NONE && index >= pr.indexOffset {
		node, err := pr.getNode(index)
		if err != nil {
			return nil, err
		}
		chain = append(chain, ExtendedRef{node.Ref.Slot, node.Ref.Root, node.ParentRoot})
		index = node.Parent
	}
	return chain, nil
}

// Returns the closest empty-slot node to the given slot. Nodes with blocks after the anchor are ignored.
func (pr *ProtoArray) ClosestToSlot(anchor Root, slot Slot) (closest NodeRef, err error) {
	max := NodeRef{Root: anchor, Slot: slot}
	// If it just exists already, we have nothing to do
	if _, ok := pr.indices[max]; ok {
		return max, nil
	}

	anchorSlot, ok := pr.blockSlots[anchor]
	if !ok {
		return NodeRef{}, fmt.Errorf("unknown anchor %s", anchor)
	}
	if anchorSlot > slot {
		return NodeRef{}, fmt.Errorf("cannot look for slot %d before anchor slot %d (%s)",
			slot, anchorSlot, anchor)
	}
	min := NodeRef{Root: anchor, Slot: anchorSlot}
	// short-cut common case: the slot is the anchor itself, e.g. we're building on the requested node.
	if anchorSlot == slot {
		return min, nil
	}

	// binary search the existing nodes for the slot closest by, with the same root.
	pivot := NodeRef{Slot: 0, Root: anchor}
	for min.Slot+1 < max.Slot {
		// max.Slot is always at least 2 higher, we're changing the pivot and don't get stuck.
		pivot.Slot = min.Slot + ((max.Slot - min.Slot) / 2)
		if _, ok := pr.indices[pivot]; ok {
			min.Slot = pivot.Slot
		} else {
			max.Slot = pivot.Slot
		}
	}
	return min, nil
}

// Returns the canonical node at the given slot.
// If preBlock is true, a slot node is retrieved. If false, a block node is retrieved, or nil if the slot is empty.
// If there is a double proposal, the canonical block (w.r.t. votes) is used.
// If the fork-choice starts at a filled slot node, this node cannot be requested with preBlock == true,
// as the data is already pruned.
func (pr *ProtoArray) CanonAtSlot(anchor Root, slot Slot, preBlock bool) (closest *NodeRef, err error) {
	anchorSlot, ok := pr.blockSlots[anchor]
	if !ok {
		return nil, fmt.Errorf("unknown anchor %s", anchor)
	}
	if anchorSlot > slot {
		return nil, fmt.Errorf("cannot look for slot %d before anchor slot %d (%s)",
			slot, anchorSlot, anchor)
	}
	// short-cut common case: the slot is the anchor itself, e.g. we're building on the current head.
	if anchorSlot == slot {
		ref := NodeRef{Root: anchor, Slot: slot}
		if preBlock {
			i, ok := pr.indices[ref]
			if !ok {
				panic("anchor node is missing")
			}
			node := &pr.nodes[i]
			// Is the anchor a filled node?
			if node.ParentRoot != anchor {
				return nil, fmt.Errorf("cannot look for slot %d at anchor, anchor is filled node", slot)
			}
		}
		return &ref, nil
	}
	head, err := pr.FindHead(anchor, anchorSlot)
	if err != nil {
		return nil, err
	}
	// The head may be the closest we have.
	if head.Slot <= slot {
		return &head, nil
	}
	// Walk back the canonical chain, and stop as soon as we find the node at slot of interest.
	index := pr.indices[head]
	var node *ProtoNode
	for index != NONE && index >= pr.indexOffset {
		node, err = pr.getNode(index)
		if err != nil {
			return nil, err
		}
		// if we are looking for the pre-block node, and the node is not empty, then skip it.
		if preBlock && node.ParentRoot != node.Ref.Root {
			index = node.Parent
			continue
		}
		if node.Ref.Slot == slot {
			if !preBlock && node.Ref.Root == node.ParentRoot {
				// no block exists for this slot, it's empty.
				return nil, nil
			}
			res := node.Ref
			return &res, nil
		}
		if node.Ref.Slot < slot {
			break // should never happen (a missing gap slot node), but handle gracefully
		}
		index = node.Parent
	}
	return nil, fmt.Errorf("cannot find node at slot %d (pre block %v)", slot, preBlock)
}

func (pr *ProtoArray) GetNode(blockRoot Root) (NodeRef, bool) {
	slot, ok := pr.blockSlots[blockRoot]
	if !ok {
		return NodeRef{}, false
	}
	index, ok := pr.indices[NodeRef{Root: blockRoot, Slot: slot}]
	if !ok {
		panic("indices is inconsistent with blockSlots")
	}
	node, err := pr.getNode(index)
	if err != nil {
		panic("indices is inconsistent with getNode")
	}
	return node.Ref, true
}

var lengthMismatchErr = errors.New("length mismatch")

// Iterate backwards through the array, touching all nodes and their parents and potentially
// the best-child of each parent.
//
// The structure of the `self.nodes` array ensures that the child of each node is always
// touched before its parent.
//
// For each node, the following is done:
//
// - Update the node's weight with the corresponding delta (can be negative).
// - Back-propagate each node's delta to its parents delta.
// - Compare the current node with the parents best-child, updating it if the current node
// should become the best child.
// - If required, update the parents best-descendant with the current node or its best-descendant.
func (pr *ProtoArray) ApplyScoreChanges(deltas []SignedGwei, justifiedEpoch Epoch, finalizedEpoch Epoch) error {
	if len(deltas) != len(pr.nodes) {
		return lengthMismatchErr
	}
	if justifiedEpoch != pr.justifiedEpoch || finalizedEpoch != pr.finalizedEpoch {
		pr.justifiedEpoch = justifiedEpoch
		pr.finalizedEpoch = finalizedEpoch
	}
	for i := len(pr.nodes) - 1; i >= 0; i-- {
		delta := deltas[i]
		node := &pr.nodes[i]
		node.Weight += delta
		if node.Parent != NONE {
			deltas[node.Parent-pr.indexOffset] += delta
		}
	}
	for i := len(pr.nodes) - 1; i >= 0; i-- {
		node := &pr.nodes[i]
		if node.Parent != NONE {
			if err := pr.maybeUpdateBestChildAndDescendant(node.Parent, pr.indexOffset+ProtoNodeIndex(i)); err != nil {
				return err
			}
		}
	}
	pr.updatedConnections = true
	return nil
}

func (pr *ProtoArray) updateConnections() error {
	for i := len(pr.nodes) - 1; i >= 0; i-- {
		node := &pr.nodes[i]
		if node.Parent != NONE {
			if err := pr.maybeUpdateBestChildAndDescendant(node.Parent, pr.indexOffset+ProtoNodeIndex(i)); err != nil {
				return err
			}
		}
	}
	pr.updatedConnections = true
	return nil
}

// Called to add an empty slot to the graph.
// Any gaps between the existing graph nodes and the given slot are filled with nodes.
// If the justifiedEpoch or finalizedEpoch changed in-between the existing graph and the new slot,
// the call should be split with increasing slot numbers, to reflect the changes of justification and finalization.
func (pr *ProtoArray) OnSlot(parent Root, slot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) {
	nodeRef := NodeRef{Root: parent, Slot: slot}
	// If the node is already known, simply ignore it.
	_, ok := pr.indices[nodeRef]
	if ok {
		return
	}
	parentIndex := NONE
	parentSlot, ok := pr.blockSlots[parent]
	if ok {
		for i := parentSlot; i < slot; i++ {
			nodeRef := NodeRef{Root: parent, Slot: i}
			// remember the last node before (up to and including same slot)
			nodeIndex, ok := pr.indices[nodeRef]
			// Does this node already exist? (e.g. a branch of nodes with empty slots). Skip if so.
			if ok {
				parentIndex = nodeIndex
				continue
			}
			// No node to represent space between parent slot and new slot yet, so we add it.
			nodeIndex = pr.indexOffset + ProtoNodeIndex(len(pr.nodes))
			pr.indices[nodeRef] = nodeIndex
			pr.nodes = append(pr.nodes, ProtoNode{
				Ref:            nodeRef,
				Parent:         parentIndex,
				ParentRoot:     parent,
				JustifiedEpoch: justifiedEpoch,
				FinalizedEpoch: finalizedEpoch,
				Weight:         0,
				BestChild:      NONE,
				BestDescendant: NONE,
			})
			// remember the node as parent for the next
			parentIndex = nodeIndex
		}
	}
	// Add the node for the slot
	nodeIndex := pr.indexOffset + ProtoNodeIndex(len(pr.nodes))
	pr.indices[nodeRef] = nodeIndex
	pr.nodes = append(pr.nodes, ProtoNode{
		Ref:            nodeRef,
		Parent:         parentIndex,
		ParentRoot:     parent,
		JustifiedEpoch: justifiedEpoch,
		FinalizedEpoch: finalizedEpoch,
		Weight:         0,
		BestChild:      NONE,
		BestDescendant: NONE,
	})
	// Connections are out of sync, i.e. array needs work before next find-head can return the proper head.
	pr.updatedConnections = false
}

// Register a block with the fork choice. Calls OnSlot to add any missing slot nodes.
// If justified or finalized in-between, make sure to call OnSlot with accurate details first.
//
// The parent root of the genesis block should be zeroed.
func (pr *ProtoArray) OnBlock(parent Root, blockRoot Root, blockSlot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) {
	blockRef := NodeRef{Root: blockRoot, Slot: blockSlot}
	// If the block is already known, simply ignore it.
	_, ok := pr.indices[blockRef]
	if ok {
		return
	}
	if _, ok := pr.blockSlots[blockRoot]; ok {
		panic("cannot add block with same root but different slot")
	}
	pr.OnSlot(parent, blockSlot, justifiedEpoch, finalizedEpoch)
	exclRef := NodeRef{Slot: blockSlot, Root: parent}
	parentIndex, ok := pr.indices[exclRef]
	if !ok {
		panic("OnSlot failed to add node for block slot")
	}
	nodeIndex := pr.indexOffset + ProtoNodeIndex(len(pr.nodes))
	pr.blockSlots[blockRoot] = blockSlot
	pr.indices[blockRef] = nodeIndex
	pr.nodes = append(pr.nodes, ProtoNode{
		Ref:            blockRef,
		Parent:         parentIndex,
		ParentRoot:     parent,
		JustifiedEpoch: justifiedEpoch,
		FinalizedEpoch: finalizedEpoch,
		Weight:         0,
		BestChild:      NONE,
		BestDescendant: NONE,
	})
	// Connections are out of sync, i.e. array needs work before next find-head can return the proper head.
	pr.updatedConnections = false
}

var UnknownAnchorErr = errors.New("anchor unknown")
var NoViableHeadErr = errors.New("not a viable head anymore, invalid forkchoice state")

// Finds the head, starting from the anchor_root subtree. (justified_root for regular fork-choice)
//
// Follows the best-descendant links to find the best-block (i.e., head-block).
//
// The result of this function is not guaranteed to be accurate if `OnBlock` has
// been called without a subsequent `applyScoreChanges` call. This is because
// `OnBlock` does not attempt to walk backwards through the tree and update the
// best-child/best-descendant links.
func (pr *ProtoArray) FindHead(anchorRoot Root, anchorSlot Slot) (NodeRef, error) {
	if !pr.updatedConnections {
		if err := pr.updateConnections(); err != nil {
			return NodeRef{}, err
		}
	}
	anchorRef := NodeRef{Root: anchorRoot, Slot: anchorSlot}
	anchorIndex, ok := pr.indices[anchorRef]
	if !ok {
		return NodeRef{}, UnknownAnchorErr
	}
	anchorNode, err := pr.getNode(anchorIndex)
	if err != nil {
		return NodeRef{}, err
	}
	bestDescIndex := anchorNode.BestDescendant
	if bestDescIndex == NONE {
		bestDescIndex = anchorIndex
	}
	bestNode, err := pr.getNode(bestDescIndex)
	if err != nil {
		return NodeRef{}, err
	}
	if !pr.isNodeViableForHead(bestNode) {
		return NodeRef{}, NoViableHeadErr
	}
	return bestNode.Ref, nil
}

// IsAncestor checks if root is an ancestor of ofRoot. Equal roots do not count as ancestor.
func (pr *ProtoArray) IsAncestor(root Root, ofRoot Root) (unknown bool, isAncestor bool) {
	// can't be ancestors if they are equal.
	if root == ofRoot {
		return false, false
	}
	if !pr.updatedConnections {
		if err := pr.updateConnections(); err != nil {
			return true, false
		}
	}
	ofSlot, ok := pr.blockSlots[ofRoot]
	if !ok {
		return true, false
	}
	anchorRef := NodeRef{Root: ofRoot, Slot: ofSlot}
	anchorIndex, ok := pr.indices[anchorRef]
	if !ok {
		return true, false
	}
	anchorNode, err := pr.getNode(anchorIndex)
	if err != nil {
		return true, false
	}
	slot, ok := pr.blockSlots[root]
	if !ok {
		return true, false
	}
	lookupRef := NodeRef{Root: root, Slot: slot}
	lookupIndex, ok := pr.indices[lookupRef]
	if !ok {
		return true, false
	}
	lookupNode, err := pr.getNode(lookupIndex)
	if err != nil {
		return true, false
	}
	// ofRoot is later on the same chain than the looked up root.
	// So ofRoot may be an ancestor of root, but not vice versa.
	if anchorNode.Ref.Slot >= lookupNode.Ref.Slot {
		return false, false
	}
	sameChain := anchorNode.BestDescendant == lookupNode.BestDescendant
	return false, sameChain
}

var HeadUnknownErr = errors.New("array has invalid state, head has no index")

type prunedNode struct {
	canonical bool
	node      *ProtoNode
}

// Update the tree with new finalization information (or alternatively another trusted root and slot)
// The slot may point to a gap slot,
// in which case the node with the anchor block of the anchor block-root is pruned,
// and the next nodes, up to (and excl.) the anchorSlot.
func (pr *ProtoArray) OnPrune(anchorRoot Root, anchorSlot Slot) error {
	anchorRef := NodeRef{Root: anchorRoot, Slot: anchorSlot}
	anchorIndex, ok := pr.indices[anchorRef]
	if !ok {
		// if the anchor is unknown, then there is nothing to prune anyway.
		return nil
	}
	if anchorIndex == pr.indexOffset {
		// nothing to do
		return nil
	}
	// Get the head, it will help quickly determine if pruned nodes are canonical
	head, err := pr.FindHead(anchorRoot, anchorSlot)
	if err != nil {
		return err
	}
	headIndex, ok := pr.indices[head]
	if !ok {
		return HeadUnknownErr
	}
	// Remove the `self.indices` and `self.blockSlots` key/values for all the to-be-deleted nodes.
	j := 0
	var pruned []prunedNode
	for i := pr.indexOffset; i < anchorIndex; i++ {
		node := &pr.nodes[j]
		if pr.sink != nil {
			canonical := node.BestDescendant == headIndex
			pruned = append(pruned, prunedNode{canonical, node})
		}
		// Remove one by one (oldest first), so above errors cannot mess up forkchoice state
		delete(pr.indices, node.Ref)
		// Remove the block-slots ref
		delete(pr.blockSlots, node.Ref.Root)
		// TODO: is this slicing bad for GC?
		pr.nodes = pr.nodes[1:]
		// update offset
		pr.indexOffset = i
	}
	// adjust the slot we know for the anchor root, everything before it was pruned.
	pr.blockSlots[anchorRoot] = anchorSlot
	// Send pruned nodes to the node sink (empty if no sink)
	for _, p := range pruned {
		if err := pr.sink.OnPrunedNode(p.node.Ref, p.canonical); err != nil {
			return err
		}
	}
	return nil
}

// Observe the parent at `parent_index` with respect to the child at `child_index` and
// potentially modify the `parent.best_child` and `parent.best_descendant` values.
//
// There are four outcomes:
//
// - The child is already the best child but it's now invalid due to a FFG change and should be removed.
// - The child is already the best child and the parent is updated with the new best-descendant.
// - The child is not the best child but becomes the best child.
// - The child is not the best child and does not become the best child.
func (pr *ProtoArray) maybeUpdateBestChildAndDescendant(parentIndex ProtoNodeIndex, childIndex ProtoNodeIndex) error {
	child, err := pr.getNode(childIndex)
	if err != nil {
		return err
	}
	parent, err := pr.getNode(parentIndex)
	if err != nil {
		return err
	}
	childLeadsToViableHead, err := pr.nodeLeadsToViableHead(child)
	if err != nil {
		return err
	}

	changeToNone := func() {
		parent.BestChild = NONE
		parent.BestDescendant = NONE
	}

	changeToChild := func() {
		parent.BestChild = childIndex
		if child.BestDescendant == NONE {
			parent.BestDescendant = childIndex
		} else {
			parent.BestDescendant = child.BestDescendant
		}
	}

	if parent.BestChild != NONE {
		if parent.BestChild == childIndex {
			if !childLeadsToViableHead {
				// If the child is already the best-child of the parent but it's not viable for the head, remove it.
				changeToNone()
			} else {
				// If the child is the best-child already, set it again to ensure that the
				// best-descendant of the parent is updated.
				changeToChild()
			}
		} else {
			bestChild, err := pr.getNode(parent.BestChild)
			if err != nil {
				return err
			}
			bestChildLeadsToViableHead, err := pr.nodeLeadsToViableHead(bestChild)
			if err != nil {
				return err
			}

			if childLeadsToViableHead && !bestChildLeadsToViableHead {
				// The child leads to a viable head, but the current best-child doesn't.
				changeToChild()
			} else if (!childLeadsToViableHead) && bestChildLeadsToViableHead {
				// The best child leads to a viable head, but the child doesn't.
				// *No change*
			} else if child.Weight == bestChild.Weight {
				// Tie-breaker of equal weights by root.
				for i := 0; i < 32; i++ {
					if child.Ref.Root[i] >= bestChild.Ref.Root[i] {
						changeToChild()
						break
					}
				}
				// otherwise *no change*
			} else {
				// Choose the winner by weight.
				if child.Weight >= bestChild.Weight {
					changeToChild()
				}
				// otherwise *no change*
			}
		}
	} else {
		if childLeadsToViableHead {
			// There is no current best-child and the child is viable.
			changeToChild()
		} else {
			// There is no current best-child but the child is not viable.
			// *No change*
		}
	}
	return nil
}

// Indicates if the node itself is viable for the head, or if it's best descendant is viable for the head.
func (pr *ProtoArray) nodeLeadsToViableHead(node *ProtoNode) (bool, error) {
	if node.BestDescendant != NONE {
		best, err := pr.getNode(node.BestDescendant)
		if err != nil {
			return false, err
		}
		return pr.isNodeViableForHead(best), nil
	} else {
		return pr.isNodeViableForHead(node), nil
	}
}

//This is the equivalent to the `filter_block_tree` function in the eth2 spec:
//
//https://github.com/ethereum/eth2.0-specs/blob/v0.11.1/specs/phase0/fork-choice.md#filter_block_tree
//
//Any node that has a different finalized or justified epoch should not be viable for the head.
func (pr *ProtoArray) isNodeViableForHead(node *ProtoNode) bool {
	return (node.JustifiedEpoch == pr.justifiedEpoch || pr.justifiedEpoch == beacon.GENESIS_EPOCH) &&
		(node.FinalizedEpoch == pr.finalizedEpoch || pr.finalizedEpoch == beacon.GENESIS_EPOCH)
}

type VoteTracker struct {
	Current            NodeRef
	Next               NodeRef
	CurrentTargetEpoch Epoch
	NextTargetEpoch    Epoch
}

type ForkChoice struct {
	sync.RWMutex
	protoArray *ProtoArray
	votes      []VoteTracker
	balances   []Gwei
	// The block root to start forkchoice from.
	// At genesis, explicitly set to genesis, since there is no "justified" root yet.
	// Later updated with justified state.
	// May be modified to pin a sub-tree for soft-fork purposes.
	anchor    NodeRef
	justified Checkpoint
	finalized Checkpoint
	spec      *beacon.Spec
}

func NewForkChoice(spec *beacon.Spec, finalized Checkpoint, justified Checkpoint,
	anchorRoot Root, anchorSlot Slot, anchorParent Root, sink NodeSink) *ForkChoice {
	fc := &ForkChoice{
		protoArray: NewProtoArray(justified.Epoch, finalized.Epoch, sink),
		votes:      nil,
		balances:   nil,
		justified:  justified,
		finalized:  finalized,
		spec:       spec,
	}
	fc.ProcessBlock(anchorParent, anchorRoot, anchorSlot, justified.Epoch, finalized.Epoch)
	fc.anchor = NodeRef{Root: anchorRoot, Slot: anchorSlot}
	return fc
}

func (fc *ForkChoice) CanonicalChain(anchorRoot Root, anchorSlot Slot) ([]ExtendedRef, error) {
	fc.Lock()
	defer fc.Unlock()
	return fc.CanonicalChain(anchorRoot, anchorSlot)
}

// Process an attestation. (Note that the head slot may be for a gap slot after the block root)
func (fc *ForkChoice) ProcessAttestation(index ValidatorIndex, blockRoot Root, headSlot Slot) {
	fc.Lock()
	defer fc.Unlock()
	if index >= ValidatorIndex(len(fc.votes)) {
		if index < ValidatorIndex(cap(fc.votes)) {
			fc.votes = fc.votes[:index+1]
		} else {
			extension := make([]VoteTracker, index+1-ValidatorIndex(len(fc.votes)))
			fc.votes = append(fc.votes, extension...)
		}
	}
	vote := &fc.votes[index]
	targetEpoch := fc.spec.SlotToEpoch(headSlot)
	// only update if it's a newer vote
	if targetEpoch > vote.NextTargetEpoch {
		vote.NextTargetEpoch = targetEpoch
		vote.Next = NodeRef{Root: blockRoot, Slot: headSlot}
	}
	// TODO: maybe help detect slashable votes on the fly?
}

func (fc *ForkChoice) ProcessSlot(parentRoot Root, slot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) {
	fc.Lock()
	defer fc.Unlock()
	fc.protoArray.OnSlot(parentRoot, slot, justifiedEpoch, finalizedEpoch)
}

func (fc *ForkChoice) ProcessBlock(parentRoot Root, blockRoot Root, blockSlot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) {
	fc.Lock()
	defer fc.Unlock()
	fc.protoArray.OnBlock(parentRoot, blockRoot, blockSlot, justifiedEpoch, finalizedEpoch)
}

func (fc *ForkChoice) PinAnchor(root Root, slot Slot) error {
	fc.Lock()
	defer fc.Unlock()
	if fc.anchor.Root == root {
		return nil
	}
	if unknown, isAncestor := fc.IsAncestor(root, fc.anchor.Root); unknown {
		return fmt.Errorf("missing anchor data, cannot change anchor from (%s, %d) to (%s, %d)",
			fc.anchor.Root, fc.anchor.Slot, root, slot)
	} else if !isAncestor {
		return fmt.Errorf("cannot pin anchor to (%s, %d) which is not in the subtree of the previous anchor (%s, %d)",
			root, slot, fc.anchor.Root, fc.anchor.Slot)
	}
	fc.anchor = NodeRef{Root: root, Slot: slot}
	return nil
}

// UpdateJustified updates what is recognized as justified and finalized checkpoint,
// and adjusts justified balances for vote weights.
// If the finalized checkpoint changes, it triggers pruning.
// Note that pruning can prune the pre-block node of the start slot of the finalized epoch, if it is not a gap slot.
// And the finalizing node with the block will remain.
func (fc *ForkChoice) UpdateJustified(justified Checkpoint, finalized Checkpoint, justifiedStateBalances []Gwei) error {
	fc.Lock()
	defer fc.Unlock()
	justifiedSlot, err := fc.spec.EpochStartSlot(justified.Epoch)
	if err != nil {
		return fmt.Errorf("bad justified epoch: %d", justified.Epoch)
	}
	if justified.Epoch < finalized.Epoch {
		return fmt.Errorf("justified epoch %d lower than finalized epoch %d", justified.Epoch, finalized.Epoch)
	}

	oldBals := fc.balances
	newBals := justifiedStateBalances

	deltas := computeDeltas(fc.protoArray.indices, fc.votes, oldBals, newBals)

	if err := fc.protoArray.ApplyScoreChanges(deltas, justified.Epoch, finalized.Epoch); err != nil {
		return err
	}

	fc.balances = newBals
	fc.justified = justified
	prevFinalized := fc.finalized
	fc.finalized = finalized
	fc.anchor = NodeRef{Root: justified.Root, Slot: justifiedSlot}

	// prune if we finalized something
	if prevFinalized != finalized {
		finSlot, _ := fc.spec.EpochStartSlot(finalized.Epoch)
		if err := fc.protoArray.OnPrune(finalized.Root, finSlot); err != nil {
			return err
		}
	}
	return nil
}

func (fc *ForkChoice) Justified() Checkpoint {
	return fc.justified
}

func (fc *ForkChoice) Finalized() Checkpoint {
	return fc.finalized
}

func (fc *ForkChoice) IsAncestor(root Root, ofRoot Root) (unknown bool, isAncestor bool) {
	fc.Lock()
	defer fc.Unlock()
	return fc.protoArray.IsAncestor(root, ofRoot)
}

func (fc *ForkChoice) ClosestToSlot(anchor Root, slot Slot) (ref NodeRef, err error) {
	fc.Lock()
	defer fc.Unlock()
	return fc.protoArray.ClosestToSlot(anchor, slot)
}

func (fc *ForkChoice) CanonAtSlot(anchor Root, slot Slot, preBlock bool) (closest *NodeRef, err error) {
	fc.Lock()
	defer fc.Unlock()
	return fc.protoArray.CanonAtSlot(anchor, slot, preBlock)
}

// Return the latest block reference known for the node.
// Warning: if it is a gap slot there may be earlier nodes with the same root, but pruned.
func (fc *ForkChoice) GetNode(root Root) (ref NodeRef, ok bool) {
	fc.RLock()
	defer fc.RUnlock()
	return fc.protoArray.GetNode(root)
}

func (fc *ForkChoice) FindHead() (NodeRef, error) {
	fc.Lock()
	defer fc.Unlock()
	return fc.protoArray.FindHead(fc.anchor.Root, fc.anchor.Slot)
}

// Returns a list of `deltas`, where there is one delta for each of the ProtoArray nodes.
// The deltas are calculated between `oldBalances` and `newBalances`, and/or a change of vote.
func computeDeltas(indices map[NodeRef]ProtoNodeIndex, votes []VoteTracker, oldBalances []Gwei, newBalances []Gwei) []SignedGwei {
	deltas := make([]SignedGwei, len(indices), len(indices))
	for i := 0; i < len(votes); i++ {
		vote := &votes[i]
		// There is no need to create a score change if the validator has never voted (may not be active)
		// or both their votes are for the zero checkpoint (alias to the genesis block).
		if vote.Current == (NodeRef{}) && vote.Next == (NodeRef{}) {
			continue
		}

		// Validator sets may have different sizes (but attesters are not different, activation only under finality)
		oldBal := Gwei(0)
		if i < len(oldBalances) {
			oldBal = oldBalances[i]
		}
		newBal := Gwei(0)
		if i < len(newBalances) {
			newBal = newBalances[i]
		}

		if vote.CurrentTargetEpoch < vote.NextTargetEpoch || oldBal != newBal {
			// Ignore the current or next vote if it is not known in `indices`.
			// We assume that it is outside of our tree (i.e., pre-finalization) and therefore not interesting.
			if currentIndex, ok := indices[vote.Current]; ok {
				deltas[currentIndex] -= SignedGwei(oldBal)
			}
			if nextIndex, ok := indices[vote.Next]; ok {
				deltas[nextIndex] += SignedGwei(newBal)
			}
			vote.Current = vote.Next
			vote.CurrentTargetEpoch = vote.NextTargetEpoch
		}
	}

	return deltas
}
