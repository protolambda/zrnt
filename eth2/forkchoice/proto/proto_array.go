package proto

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	. "github.com/protolambda/zrnt/eth2/forkchoice"
)

const NONE = ^NodeIndex(0)

type ProtoNode struct {
	Ref NodeRef
	// The transition parent of a block node is the node of the same slot, without the block.
	// The transition parent of a slot node is the slot or block before it.
	TransitionParent NodeIndex
	// The forkchoice parent of a node is strictly one slot lower, it cannot be the same slot.
	ForkchoiceParent NodeIndex
	// Duplicated to avoid pruning of this useful info.
	ParentRoot     Root
	JustifiedEpoch Epoch
	FinalizedEpoch Epoch
	Weight         SignedGwei
	// Relative to ForkchoiceParent relations
	BestChild NodeIndex
	// Relative to ForkchoiceParent relations
	BestDescendant NodeIndex
}

type NodeSinkFn func(ctx context.Context, ref NodeRef, canonical bool) error

func (fn NodeSinkFn) OnPrunedNode(ctx context.Context, ref NodeRef, canonical bool) error {
	return fn(ctx, ref, canonical)
}

type NodeSink interface {
	OnPrunedNode(ctx context.Context, ref NodeRef, canonical bool) error
}

// Tracks slots and blocks as nodes.
// Every block has two nodes: with and without the block. The node with the block is the child of that without it.
// Gap slots just have a single node.
// There may be multiple nodes with the same parent but different blocks (i.e. double proposals, but slashable).
type ProtoArray struct {
	sink           NodeSink
	indexOffset    NodeIndex
	justifiedEpoch Epoch
	finalizedEpoch Epoch
	nodes          []ProtoNode
	// maintains only nodes that are actually part of the tree starting from finalized point.
	indices map[NodeRef]NodeIndex
	// Tracks the first slot at or after the block root that the array knows of.
	// The lowest slot for a block does not equal the block.slot itself, that may have been pruned.
	blockSlots         map[Root]Slot
	updatedConnections bool
}

var _ ForkchoiceGraph = (*ProtoArray)(nil)

func NewProtoArray(parent Root, blockRoot Root, blockSlot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch, sink NodeSink) *ProtoArray {
	blockRef := NodeRef{Root: blockRoot, Slot: blockSlot}
	pr := ProtoArray{
		sink:               sink,
		indexOffset:        0,
		justifiedEpoch:     justifiedEpoch,
		finalizedEpoch:     finalizedEpoch,
		nodes:              make([]ProtoNode, 0, 100),
		indices:            make(map[NodeRef]NodeIndex, 100),
		blockSlots:         make(map[Root]Slot, 100),
		updatedConnections: true,
	}
	pr.blockSlots[blockRoot] = blockSlot
	pr.indices[blockRef] = 0
	pr.nodes = append(pr.nodes, ProtoNode{
		Ref:              blockRef,
		TransitionParent: NONE,
		ForkchoiceParent: NONE,
		ParentRoot:       parent,
		JustifiedEpoch:   justifiedEpoch,
		FinalizedEpoch:   finalizedEpoch,
		Weight:           0,
		BestChild:        NONE,
		BestDescendant:   NONE,
	})
	return &pr
}

var invalidIndexErr = errors.New("invalid index")

func (pr *ProtoArray) getNode(index NodeIndex) (*ProtoNode, error) {
	if index < pr.indexOffset {
		return nil, invalidIndexErr
	}
	i := index - pr.indexOffset
	if i > NodeIndex(len(pr.nodes)) {
		return nil, invalidIndexErr
	}
	return &pr.nodes[i], nil
}

func (pr *ProtoArray) Indices() map[NodeRef]NodeIndex {
	return pr.indices
}

// From head back to anchor root (including the anchor itself, if present) and anchor slot.
// Includes nodes with empty block, then followed up by a node with the block if there is any.
func (pr *ProtoArray) CanonicalChain(anchorRoot Root, anchorSlot Slot) ([]ExtendedNodeRef, error) {
	head, err := pr.FindHead(anchorRoot, anchorSlot)
	if err != nil {
		return nil, err
	}
	chain := make([]ExtendedNodeRef, 0, len(pr.nodes))
	index := pr.indices[head]
	for index != NONE && index >= pr.indexOffset {
		node, err := pr.getNode(index)
		if err != nil {
			return nil, err
		}
		chain = append(chain, ExtendedNodeRef{NodeRef: node.Ref, ParentRoot: node.ParentRoot})
		index = node.TransitionParent
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
// If withBlock is false, a slot node is retrieved. If true, a block node is retrieved, or nil if the slot is empty.
// If there is a double proposal, the canonical block (w.r.t. votes) is used.
// If the fork-choice starts at a filled slot node, this node cannot be requested with withBlock == false,
// as the data is already pruned.
func (pr *ProtoArray) CanonAtSlot(anchor Root, slot Slot, withBlock bool) (at NodeRef, err error) {
	anchorSlot, ok := pr.blockSlots[anchor]
	if !ok {
		return NodeRef{}, fmt.Errorf("unknown anchor %s", anchor)
	}
	if anchorSlot > slot {
		return NodeRef{}, fmt.Errorf("cannot look for slot %d before anchor slot %d (%s)",
			slot, anchorSlot, anchor)
	}
	// short-cut common case: the slot is the anchor itself, e.g. we're building on the current head.
	if anchorSlot == slot {
		ref := NodeRef{Root: anchor, Slot: slot}
		if !withBlock {
			i, ok := pr.indices[ref]
			if !ok {
				panic("anchor node is missing")
			}
			node := &pr.nodes[i]
			// Is the anchor a filled node?
			if node.ParentRoot != anchor {
				return NodeRef{}, fmt.Errorf("cannot look for pre-block %d at anchor, anchor is post-block", slot)
			}
		}
		return ref, nil
	}
	head, err := pr.FindHead(anchor, anchorSlot)
	if err != nil {
		return NodeRef{}, err
	}
	// The head may be the closest we have.
	if head.Slot <= slot {
		return head, nil
	}
	// Walk back the canonical chain, and stop as soon as we find the node at slot of interest.
	index := pr.indices[head]
	var node *ProtoNode
	for index != NONE && index >= pr.indexOffset {
		node, err = pr.getNode(index)
		if err != nil {
			return NodeRef{}, err
		}
		// if we are looking for the pre-block node, and the node is not empty, then skip it.
		if !withBlock && node.ParentRoot != node.Ref.Root {
			index = node.TransitionParent
			continue
		}
		if node.Ref.Slot == slot {
			if withBlock && node.Ref.Root == node.ParentRoot {
				// no block exists for this slot, it's empty.
				return NodeRef{}, nil
			}
			return node.Ref, nil
		}
		if node.Ref.Slot < slot {
			break // should never happen (a missing gap slot node), but handle gracefully
		}
		index = node.TransitionParent
	}
	return NodeRef{}, fmt.Errorf("cannot find node at slot %d (with block %v)", slot, withBlock)
}

func (pr *ProtoArray) GetSlot(blockRoot Root) (Slot, bool) {
	slot, ok := pr.blockSlots[blockRoot]
	return slot, ok
}

// Searches the available nodes for blocks with a matching parent root and/or matching slot.
// If no options are specified, the
func (pr *ProtoArray) Search(anchor NodeRef, parentRoot *Root, slot *Slot) (nonCanon []NodeRef, canon []NodeRef, err error) {
	// this also checks that the anchor exists and updates the node connections.
	head, err := pr.FindHead(anchor.Root, anchor.Slot)
	if err != nil {
		return nil, nil, err
	}
	anchorIndex := pr.indices[anchor]
	headIndex := pr.indices[head]
	for i := 0; i < len(pr.nodes); i++ {
		node := &pr.nodes[i]
		// only search for nodes that contain blocks
		if node.Ref.Root == node.ParentRoot {
			continue
		}
		// no options = search for heads.
		if parentRoot == nil && slot == nil {
			// if it has no child, it's a head.
			if node.BestChild != NONE {
				// if it has only empty slots as children, it's a head.
				desc := &pr.nodes[node.BestDescendant]
				if desc.Ref.Root != node.Ref.Root {
					continue
				}
			}
		} else {
			if parentRoot != nil && node.ParentRoot != *parentRoot {
				continue
			}
			if slot != nil && node.Ref.Slot != *slot {
				continue
			}
		}
		// only output nodes that are within view (slow-ish, but does not run often thanks to above filter.)
		index := pr.indices[node.Ref]
		if _, inSubtree := pr.inSubtree(anchorIndex, index); !inSubtree {
			continue
		}
		// if it is the head, or if it has the same best descendant as the head, it's canonical.
		if node.Ref == head || node.BestDescendant == headIndex {
			canon = append(canon, node.Ref)
		} else {
			nonCanon = append(nonCanon, node.Ref)
		}
	}
	return
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
		if node.ForkchoiceParent != NONE {
			deltas[node.ForkchoiceParent-pr.indexOffset] += delta
		}
	}
	for i := len(pr.nodes) - 1; i >= 0; i-- {
		node := &pr.nodes[i]
		if node.ForkchoiceParent != NONE {
			if err := pr.maybeUpdateBestChildAndDescendant(node.ForkchoiceParent, pr.indexOffset+NodeIndex(i)); err != nil {
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
		if node.ForkchoiceParent != NONE {
			if err := pr.maybeUpdateBestChildAndDescendant(node.ForkchoiceParent, pr.indexOffset+NodeIndex(i)); err != nil {
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
func (pr *ProtoArray) ProcessSlot(parent Root, slot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) {
	nodeRef := NodeRef{Root: parent, Slot: slot}
	// If the node is already known, simply ignore it.
	_, ok := pr.indices[nodeRef]
	if ok {
		return
	}
	parentIndex := NONE
	parentSlot, ok := pr.blockSlots[parent]
	if ok {
		parentIndex = pr.indices[NodeRef{Root: parent, Slot: parentSlot}]
		for i := parentSlot + 1; i < slot; i++ {
			nodeRef := NodeRef{Root: parent, Slot: i}
			// remember the last node before (up to and including same slot)
			nodeIndex, ok := pr.indices[nodeRef]
			// Does this node already exist? (e.g. a branch of nodes with empty slots). Skip if so.
			if ok {
				parentIndex = nodeIndex
				continue
			}
			// No node to represent space between parent slot and new slot yet, so we add it.
			nodeIndex = pr.indexOffset + NodeIndex(len(pr.nodes))
			pr.indices[nodeRef] = nodeIndex
			pr.nodes = append(pr.nodes, ProtoNode{
				Ref:              nodeRef,
				TransitionParent: parentIndex,
				ForkchoiceParent: parentIndex,
				ParentRoot:       parent,
				JustifiedEpoch:   justifiedEpoch,
				FinalizedEpoch:   finalizedEpoch,
				Weight:           0,
				BestChild:        NONE,
				BestDescendant:   NONE,
			})
			// remember the node as parent for the next
			parentIndex = nodeIndex
		}
	}
	// Add the node for the slot
	nodeIndex := pr.indexOffset + NodeIndex(len(pr.nodes))
	pr.indices[nodeRef] = nodeIndex
	pr.nodes = append(pr.nodes, ProtoNode{
		Ref:              nodeRef,
		TransitionParent: parentIndex,
		ForkchoiceParent: parentIndex,
		ParentRoot:       parent,
		JustifiedEpoch:   justifiedEpoch,
		FinalizedEpoch:   finalizedEpoch,
		Weight:           0,
		BestChild:        NONE,
		BestDescendant:   NONE,
	})
	// Connections are out of sync, i.e. array needs work before next find-head can return the proper head.
	pr.updatedConnections = false
}

// Register a block with the fork choice. Calls OnSlot to add any missing slot nodes.
// If justified or finalized in-between, make sure to call OnSlot with accurate details first.
//
// The parent root of the genesis block should be zeroed.
func (pr *ProtoArray) ProcessBlock(parent Root, blockRoot Root, blockSlot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) (ok bool) {
	blockRef := NodeRef{Root: blockRoot, Slot: blockSlot}
	// If the block is already known, simply ignore it.
	if _, ok := pr.indices[blockRef]; ok {
		return true
	}
	if _, ok := pr.blockSlots[blockRoot]; ok {
		// block is already known with different slot. Likely been pruned away, and we only know the later empty slot.
		return true
	}
	// cannot add block if parent already exists, or parent is after the block
	parentBlockSlot, ok := pr.blockSlots[parent]
	if !ok || parentBlockSlot >= blockSlot {
		return false
	}
	pr.ProcessSlot(parent, blockSlot, justifiedEpoch, finalizedEpoch)

	// If the parent node is not known, we cannot add the block.
	// Note: the forkchoice graph here only considers blocks, not any empty slots. This is legacy and might change.
	forkchoiceParentIndex, ok := pr.indices[NodeRef{Slot: parentBlockSlot, Root: parent}]
	if !ok {
		return false
	}

	transitionParentIndex, ok := pr.indices[NodeRef{Slot: blockSlot, Root: parent}]
	if !ok {
		panic("OnSlot failed to add node for block slot (transition parent)")
	}
	nodeIndex := pr.indexOffset + NodeIndex(len(pr.nodes))
	pr.blockSlots[blockRoot] = blockSlot
	pr.indices[blockRef] = nodeIndex
	pr.nodes = append(pr.nodes, ProtoNode{
		Ref:              blockRef,
		TransitionParent: transitionParentIndex,
		ForkchoiceParent: forkchoiceParentIndex,
		ParentRoot:       parent,
		JustifiedEpoch:   justifiedEpoch,
		FinalizedEpoch:   finalizedEpoch,
		Weight:           0,
		BestChild:        NONE,
		BestDescendant:   NONE,
	})
	// Connections are out of sync, i.e. array needs work before next find-head can return the proper head.
	pr.updatedConnections = false
	return true
}

var UnknownAnchorErr = errors.New("anchor unknown")
var NoViableHeadErr = errors.New("not a viable head anymore, invalid forkchoice state")

// Finds the head, starting from the anchor_root *forkchoice* subtree. (justified_root for regular fork-choice)
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

// InSubtree checks if root is in the subtree of the anchor.
// If the roots are the same, it still counts as in the subtree.
func (pr *ProtoArray) InSubtree(anchor Root, root Root) (unknown bool, inSubtree bool) {
	// equal roots count as in-subtree.
	if anchor == root {
		return false, true
	}
	if !pr.updatedConnections {
		if err := pr.updateConnections(); err != nil {
			return true, false
		}
	}
	anchorSlot, ok := pr.blockSlots[anchor]
	if !ok {
		return true, false
	}
	anchorRef := NodeRef{Root: anchor, Slot: anchorSlot}
	anchorIndex, ok := pr.indices[anchorRef]
	if !ok {
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
	return pr.inSubtree(anchorIndex, lookupIndex)
}

// InSubtree checks if lookupIndex is in the subtree of the anchorIndex.
// If the indices are the same, it still counts as in the subtree.
func (pr *ProtoArray) inSubtree(anchorIndex NodeIndex, lookupIndex NodeIndex) (unknown bool, inSubtree bool) {
	if anchorIndex == lookupIndex {
		return false, true
	}
	anchorNode, err := pr.getNode(anchorIndex)
	if err != nil {
		return true, false
	}
	lookupNode, err := pr.getNode(lookupIndex)
	if err != nil {
		return true, false
	}
	if anchorNode.Ref.Slot >= lookupNode.Ref.Slot {
		// anchor is later on the same chain than the looked up node.
		// So anchor may be in subtree of the looked up node, but not vice versa.
		return false, false
	}
	if anchorIndex >= lookupIndex {
		// anchor was inserted after looked up node.
		// So anchor may be in subtree of the looked up node, but not vice versa.
		return false, false
	}
	// shortcut: if they have the same relative head, they are on the same chain.
	if anchorNode.BestDescendant == lookupIndex || anchorNode.BestDescendant == lookupNode.BestDescendant {
		return false, true
	}
	// Root may still be on a different non-canonical branch out of the anchor.
	for i := lookupNode.TransitionParent; i != NONE && i >= anchorIndex; {
		tmp := &pr.nodes[i]
		// early exit: as soon as we find a node that has the same relative head as the anchor,
		// we know we are in-between the anchor and the head, thus in the subtree, thus an ancestor.
		if tmp.BestDescendant == anchorNode.BestDescendant {
			return false, true
		}
		i = tmp.TransitionParent
	}
	return false, false
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
func (pr *ProtoArray) OnPrune(ctx context.Context, anchorRoot Root, anchorSlot Slot) error {
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
	}
	// Send pruned nodes to the node sink (empty if no sink). Continue until it fails.
	// Only prune what we successfully sent to the sink.
	prunedUpTo := 0
	for _, p := range pruned {
		if err = pr.sink.OnPrunedNode(ctx, p.node.Ref, p.canonical); err != nil {
			break
		}
		prunedUpTo++
	}
	// adjust the slot we know for the anchor root, everything before it was pruned.
	pr.blockSlots[anchorRoot] = anchorSlot
	for _, p := range pruned[:prunedUpTo] {
		delete(pr.indices, p.node.Ref)
		// Remove the block-slots ref
		delete(pr.blockSlots, p.node.Ref.Root)
		// TODO: is this slicing bad for GC?
		pr.nodes = pr.nodes[1:]
		// update offset
		pr.indexOffset++
	}
	return err
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
func (pr *ProtoArray) maybeUpdateBestChildAndDescendant(parentIndex NodeIndex, childIndex NodeIndex) error {
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
				// Tie-breaker of equal weights by root. (smaller hash wins)
				if bytes.Compare(child.Ref.Root[:], bestChild.Ref.Root[:]) > 0 {
					changeToChild()
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
		}
		// else {
		// There is no current best-child but the child is not viable.
		// *No change*
		// }
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

// This is the equivalent to the `filter_block_tree` function in the eth2 spec:
//
// https://github.com/ethereum/eth2.0-specs/blob/v0.11.1/specs/phase0/fork-choice.md#filter_block_tree
//
// Any node that has a different finalized or justified epoch should not be viable for the head.
func (pr *ProtoArray) isNodeViableForHead(node *ProtoNode) bool {
	return (node.JustifiedEpoch == pr.justifiedEpoch || pr.justifiedEpoch == common.GENESIS_EPOCH) &&
		(node.FinalizedEpoch == pr.finalizedEpoch || pr.finalizedEpoch == common.GENESIS_EPOCH)
}
