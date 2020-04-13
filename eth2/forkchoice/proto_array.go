package forkchoice

import (
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
)

type Root = beacon.Root
type Epoch = beacon.Epoch
type Slot = beacon.Slot
type ValidatorIndex = beacon.ValidatorIndex
type Gwei = beacon.Gwei
type Checkpoint = beacon.Checkpoint

type SignedGwei = int64

type BlockRef struct {
	Slot Slot
	Root Root
}

type ProtoNodeIndex uint64

const NONE = ^ProtoNodeIndex(0)

type ProtoNode struct {
	Block          BlockRef
	Parent         ProtoNodeIndex
	JustifiedEpoch Epoch
	FinalizedEpoch Epoch
	Weight         SignedGwei
	BestChild      ProtoNodeIndex
	BestDescendant ProtoNodeIndex
}

type BlockSink interface {
	OnPrunedBlock(node *ProtoNode, canonical bool)
}

type ProtoArray struct {
	sink           BlockSink
	indexOffset    ProtoNodeIndex
	finalizedRoot  Root
	justifiedEpoch Epoch
	finalizedEpoch Epoch
	nodes          []ProtoNode
	// maintains only roots that are actually part of the tree starting from finalized point.
	indices            map[Root]ProtoNodeIndex
	updatedConnections bool
}

func NewProtoArray(justifiedEpoch Epoch, finalizedBlock BlockRef, sink BlockSink) *ProtoArray {
	finalizedEpoch := finalizedBlock.Slot.ToEpoch()
	arr := ProtoArray{
		sink:               sink,
		indexOffset:        1,
		finalizedRoot:      finalizedBlock.Root,
		justifiedEpoch:     justifiedEpoch,
		finalizedEpoch:     finalizedEpoch,
		nodes:              make([]ProtoNode, 0, beacon.SLOTS_PER_EPOCH*10),
		indices:            make(map[Root]ProtoNodeIndex, beacon.SLOTS_PER_EPOCH*10),
		updatedConnections: true,
	}
	arr.nodes = append(arr.nodes, ProtoNode{
		Block:          finalizedBlock,
		Parent:         NONE,
		JustifiedEpoch: justifiedEpoch,
		FinalizedEpoch: finalizedEpoch,
		Weight:         0,
		BestChild:      NONE,
		BestDescendant: NONE,
	})
	arr.indices[finalizedBlock.Root] = 0
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

// From head back to anchor root (including the anchor itself)
func (pr *ProtoArray) CanonicalChain(anchorRoot Root) ([]BlockRef, error) {
	head, err := pr.FindHead(anchorRoot)
	if err != nil {
		return nil, err
	}
	chain := make([]BlockRef, 0, len(pr.nodes))
	index := pr.indices[head.Root]
	for index != NONE && index >= pr.indexOffset {
		node, err := pr.getNode(index)
		if err != nil {
			return nil, err
		}
		chain = append(chain, node.Block)
		index = node.Parent
	}
	return chain, nil
}

func (pr *ProtoArray) BlocksAroundSlot(anchor Root, slot Slot) (before BlockRef, at BlockRef, after BlockRef, err error) {
	var head BlockRef
	head, err = pr.FindHead(anchor)
	if err != nil {
		return
	}
	if slot > head.Slot {
		err = fmt.Errorf("head is too old for slot. Head at %d, but request was %d", head.Slot, slot)
		return
	}
	// Walk back the canonical chain, and stop as soon as we find the blocks around the slot of interest.
	index := pr.indices[head.Root]
	var node *ProtoNode
	for index != NONE && index >= pr.indexOffset {
		node, err = pr.getNode(index)
		if err != nil {
			return
		}
		if node.Block.Slot > slot {
			after = node.Block
		}
		if node.Block.Slot == slot {
			at = node.Block
		}
		if node.Block.Slot < head.Slot {
			before = node.Block
			break
		}
		index = node.Parent
	}
	if at.Root == (Root{}) {
		err = errors.New("could not find block")
	}
	return
}

func (pr *ProtoArray) ContainsBlock(blockRoot Root) bool {
	_, ok := pr.indices[blockRoot]
	return ok
}

func (pr *ProtoArray) GetBlock(blockRoot Root) (BlockRef, bool) {
	index, ok := pr.indices[blockRoot]
	if !ok {
		return BlockRef{}, false
	}
	node, err := pr.getNode(index)
	if err != nil {
		return BlockRef{}, false
	}
	return node.Block, true
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

// Register a block with the fork choice.
//
// It is only sane to supply a `None` parent for the genesis block.
func (pr *ProtoArray) OnBlock(block BlockRef, parent Root, justifiedEpoch Epoch, finalizedEpoch Epoch) {
	// If the block is already known, simply ignore it.
	if pr.ContainsBlock(block.Root) {
		return
	}
	nodeIndex := pr.indexOffset + ProtoNodeIndex(len(pr.nodes))
	parentIndex, ok := pr.indices[parent]
	if !ok {
		parentIndex = NONE
	}
	pr.indices[block.Root] = nodeIndex
	pr.nodes = append(pr.nodes, ProtoNode{
		Block:          block,
		Parent:         parentIndex,
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
func (pr *ProtoArray) FindHead(anchorRoot Root) (BlockRef, error) {
	if !pr.updatedConnections {
		if err := pr.updateConnections(); err != nil {
			return BlockRef{}, err
		}
	}
	anchorIndex, ok := pr.indices[anchorRoot]
	if !ok {
		return BlockRef{}, UnknownAnchorErr
	}
	anchorNode, err := pr.getNode(anchorIndex)
	if err != nil {
		return BlockRef{}, err
	}
	bestDescIndex := anchorNode.BestDescendant
	if bestDescIndex == NONE {
		bestDescIndex = anchorIndex
	}
	bestNode, err := pr.getNode(bestDescIndex)
	if err != nil {
		return BlockRef{}, err
	}
	if !pr.isNodeViableForHead(bestNode) {
		return BlockRef{}, NoViableHeadErr
	}
	return bestNode.Block, nil
}

var HeadUnknownErr = errors.New("array has invalid state, head has no index")

// Update the tree with new finalization information (or alternatively another trusted root)
func (pr *ProtoArray) OnPrune(anchorRoot Root) error {
	anchorIndex, ok := pr.indices[anchorRoot]
	if !ok {
		return UnknownAnchorErr
	}
	if anchorIndex == pr.indexOffset {
		// nothing to do
		return nil
	}
	// Get the head, it will help quickly determine if pruned nodes are canonical
	head, err := pr.FindHead(anchorRoot)
	if err != nil {
		return err
	}
	headIndex, ok := pr.indices[head.Root]
	if !ok {
		return HeadUnknownErr
	}
	// Remove the `self.indices` key/values for all the to-be-deleted nodes.
	// And send the nodes to the block sink.
	j := 0
	for i := pr.indexOffset; i < anchorIndex; i++ {
		node := &pr.nodes[j]
		if pr.sink != nil {
			canonical := node.BestDescendant == headIndex
			pr.sink.OnPrunedBlock(node, canonical)
		}
		delete(pr.indices, node.Block.Root)
	}
	// Drop all the nodes prior to finalization.
	// TODO: is this slicing bad for GC?
	pr.nodes = pr.nodes[pr.indexOffset-anchorIndex:]
	// update offset
	pr.indexOffset = anchorIndex
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
					if child.Block.Root[i] >= bestChild.Block.Root[i] {
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
	CurrentRoot Root
	NextRoot    Root
	NextEpoch   Epoch
}

type ForkChoice struct {
	protoArray *ProtoArray
	votes      []VoteTracker
	balances   []Gwei
	justified  Checkpoint
	finalized  Checkpoint
}

var WrongFinalizedBlockErr = errors.New("wrong finalized block")

func NewForkChoice(finalizedBlock BlockRef, finalized Checkpoint, justified Checkpoint, sink BlockSink) (*ForkChoice, error) {
	if finalizedBlock.Slot.ToEpoch() != finalized.Epoch {
		return nil, WrongFinalizedBlockErr
	}
	return &ForkChoice{
		protoArray: NewProtoArray(justified.Epoch, finalizedBlock, sink),
		votes:      nil,
		balances:   nil,
		justified:  justified,
		finalized:  finalized,
	}, nil
}

func (fc *ForkChoice) ProcessAttestation(index ValidatorIndex, blockRoot Root, targetEpoch Epoch) {
	if index > ValidatorIndex(len(fc.votes)) {
		extension := make([]VoteTracker, index-ValidatorIndex(len(fc.votes)))
		fc.votes = append(fc.votes, extension...)
	}
	vote := &fc.votes[index]
	if targetEpoch > vote.NextEpoch {
		vote.NextRoot = blockRoot
		vote.NextEpoch = targetEpoch
	}
}

func (fc *ForkChoice) ProcessBlock(block BlockRef, parentRoot Root, justifiedEpoch Epoch, finalizedEpoch Epoch) {
	fc.protoArray.OnBlock(block, parentRoot, justifiedEpoch, finalizedEpoch)
}

func (fc *ForkChoice) UpdateJustified(justified Checkpoint, finalized Checkpoint, justifiedStateBalances []Gwei) error {
	oldBals := fc.balances
	newBals := justifiedStateBalances

	deltas := computeDeltas(fc.protoArray.indices, fc.votes, oldBals, newBals)

	if err := fc.protoArray.ApplyScoreChanges(deltas, justified.Epoch, finalized.Epoch); err != nil {
		return err
	}

	fc.balances = newBals
	fc.justified = justified
	fc.finalized = finalized

	return nil
}

func (fc *ForkChoice) Justified() Checkpoint {
	return fc.justified
}

func (fc *ForkChoice) Finalized() Checkpoint {
	return fc.finalized
}

func (fc *ForkChoice) BlocksAroundSlot(anchor Root, slot Slot) (before BlockRef, at BlockRef, after BlockRef, err error) {
	return fc.protoArray.BlocksAroundSlot(anchor, slot)
}

func (fc *ForkChoice) GetBlock(root Root) (block BlockRef, ok bool) {
	return fc.protoArray.GetBlock(root)
}

func (fc *ForkChoice) FindHead() (BlockRef, error) {
	return fc.protoArray.FindHead(fc.justified.Root)
}

// Returns a list of `deltas`, where there is one delta for each of the ProtoArray nodes.
// The deltas are calculated between `oldBalances` and `newBalances`, and/or a change of vote.
func computeDeltas(indices map[Root]ProtoNodeIndex, votes []VoteTracker, oldBalances []Gwei, newBalances []Gwei) []SignedGwei {
	deltas := make([]SignedGwei, len(indices), len(indices))
	for i := 0; i < len(votes); i++ {
		vote := &votes[i]
		// There is no need to create a score change if the validator has never voted (may not be active)
		// or both their votes are for the zero hash (alias to the genesis block).
		if vote.CurrentRoot == (Root{}) && vote.NextRoot == (Root{}) {
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

		if vote.CurrentRoot != vote.NextRoot || oldBal != newBal {
			// Ignore the current or next vote if it is not known in `indices`.
			// We assume that it is outside of our tree (i.e., pre-finalization) and therefore not interesting.
			if currentIndex, ok := indices[vote.CurrentRoot]; ok {
				deltas[currentIndex] -= SignedGwei(oldBal)
			}
			if nextIndex, ok := indices[vote.NextRoot]; ok {
				deltas[nextIndex] += SignedGwei(newBal)
			}
			vote.CurrentRoot = vote.NextRoot
		}
	}

	return deltas
}
