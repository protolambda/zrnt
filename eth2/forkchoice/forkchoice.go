package forkchoice

import (
	"context"
	"fmt"
	"sync"

	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type ProtoForkChoice struct {
	mu         sync.RWMutex
	protoArray ForkchoiceGraph
	voteStore  VoteStore

	balances []Gwei
	// If present, this overrules the forkchoice to start in this subtree,
	// instead of the justified checkpoint.
	pin       *NodeRef
	justified Checkpoint
	finalized Checkpoint
	spec      *common.Spec
}

var _ Forkchoice = (*ProtoForkChoice)(nil)

func NewForkChoice(spec *common.Spec, finalized Checkpoint, justified Checkpoint,
	anchorRoot Root, anchorSlot Slot, graph ForkchoiceGraph, votes VoteStore,
	initialBalances []Gwei) (Forkchoice, error) {
	fc := &ProtoForkChoice{
		protoArray: graph,
		voteStore:  votes,
		balances:   nil,
		justified:  justified,
		finalized:  finalized,
		spec:       spec,
	}
	if err := fc.SetPin(anchorRoot, anchorSlot); err != nil {
		return nil, err
	}
	if err := fc.updateJustified(finalized, justified, func() ([]Gwei, error) {
		return initialBalances, nil
	}); err != nil {
		return nil, err
	}
	return fc, nil
}

func (fc *ProtoForkChoice) Pin() *NodeRef {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.pin
}

func (fc *ProtoForkChoice) SetPin(root Root, slot Slot) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	// check if new data can be pinned: the node must exist, or we won't be able to find a head node ever again.
	closest, err := fc.protoArray.ClosestToSlot(root, slot)
	if err != nil {
		return fmt.Errorf("cannot find pin: %v", err)
	}
	if closest.Slot < slot {
		return fmt.Errorf("found pin target, but slot is too old: %d <> %d", closest.Slot, slot)
	}
	fc.pin = &NodeRef{Root: root, Slot: slot}
	return nil
}

// UpdateJustified updates what is recognized as justified and finalized checkpoint,
// and adjusts justified balances for vote weights.
// If the finalized checkpoint changes, it triggers pruning.
// Note that pruning can prune the pre-block node of the start slot of the finalized epoch, if it is not a gap slot.
// And the finalizing node with the block will remain.
// The justification/finalization trigger must be within the pinned subtree (if any).
func (fc *ProtoForkChoice) UpdateJustified(ctx context.Context, trigger Root, justified Checkpoint, finalized Checkpoint,
	justifiedStateBalances func() ([]Gwei, error)) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	// Old/same data? Ignore the change.
	if fc.justified.Epoch >= justified.Epoch && fc.finalized.Epoch >= finalized.Epoch {
		return nil
	}
	if fc.pin != nil && trigger != fc.pin.Root {
		// check trigger against pin, to ensure no justification/finalization of data that conflicts with the pin.
		if unknown, inSubtree := fc.InSubtree(fc.pin.Root, trigger); unknown {
			return fmt.Errorf("cannot justify/finalize with unknown trigger when forkchoice is pinned")
		} else if !inSubtree {
			return fmt.Errorf("cannot justify/finalize outside of pinned forkchoice tree")
		}
	}

	prevFinalized := fc.finalized

	if err := fc.updateJustified(justified, finalized, justifiedStateBalances); err != nil {
		return err
	}

	// prune if we finalized something, and undo the pin.
	if prevFinalized != finalized {
		fc.pin = nil
		finSlot, _ := fc.spec.EpochStartSlot(finalized.Epoch)
		if err := fc.protoArray.OnPrune(ctx, finalized.Root, finSlot); err != nil {
			return err
		}
	}
	return nil
}

func (fc *ProtoForkChoice) updateJustified(finalized Checkpoint, justified Checkpoint,
	justifiedStateBalances func() ([]Gwei, error)) error {
	if justified.Epoch < finalized.Epoch {
		return fmt.Errorf("justified epoch %d lower than finalized epoch %d", justified.Epoch, finalized.Epoch)
	}

	// check if new finalized checkpoint is valid
	if fc.finalized != finalized {
		if unknown, inSubtree := fc.InSubtree(fc.finalized.Root, finalized.Root); unknown {
			return fmt.Errorf("unknown finalized checkpoint: %s", finalized)
		} else if !inSubtree || fc.finalized.Epoch > finalized.Epoch {
			return fmt.Errorf("new finalized checkpoint %s is outside of finalized subtree: %s",
				finalized, fc.finalized)
		}
	}
	if fc.justified != justified {
		if unknown, inSubtree := fc.InSubtree(fc.finalized.Root, justified.Root); unknown {
			return fmt.Errorf("unknown justified checkpoint: %s", justified)
		} else if !inSubtree || fc.finalized.Epoch > justified.Epoch {
			return fmt.Errorf("new justified checkpoint %s is outside of finalized subtree: %s",
				justified, fc.finalized)
		}
	}

	oldBals := fc.balances
	newBals, err := justifiedStateBalances()
	if err != nil {
		return err
	}

	deltas := fc.voteStore.ComputeDeltas(fc.protoArray.Indices(), oldBals, newBals)

	if err := fc.protoArray.ApplyScoreChanges(deltas, justified.Epoch, finalized.Epoch); err != nil {
		return err
	}

	fc.balances = newBals
	fc.justified = justified
	fc.finalized = finalized

	return nil
}

// TODO: skip based on time (like rate limiting) or based on amount of changes
//
//	(if not bigger than previous difference between head-node contenders)
func (fc *ProtoForkChoice) updateVotesMaybe() error {
	if !fc.voteStore.HasChanges() {
		return nil
	}

	deltas := fc.voteStore.ComputeDeltas(fc.protoArray.Indices(), fc.balances, fc.balances)

	return fc.protoArray.ApplyScoreChanges(deltas, fc.justified.Epoch, fc.finalized.Epoch)
}

func (fc *ProtoForkChoice) Justified() Checkpoint {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.justified
}

func (fc *ProtoForkChoice) Finalized() Checkpoint {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.finalized
}

func (fc *ProtoForkChoice) ProcessAttestation(index ValidatorIndex, blockRoot Root, headSlot Slot) (ok bool) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	// only add the vote if we can. Don't add if it's not within view.
	blockSlot, ok := fc.protoArray.GetSlot(blockRoot)
	if !ok || blockSlot < headSlot {
		return false
	}
	return fc.voteStore.ProcessAttestation(index, blockRoot, headSlot)
}

func (fc *ProtoForkChoice) CanonicalChain(anchorRoot Root, anchorSlot Slot) ([]ExtendedNodeRef, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.protoArray.CanonicalChain(anchorRoot, anchorSlot)
}

func (fc *ProtoForkChoice) ProcessSlot(parentRoot Root, slot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.protoArray.ProcessSlot(parentRoot, slot, justifiedEpoch, finalizedEpoch)
}

func (fc *ProtoForkChoice) ProcessBlock(parentRoot Root, blockRoot Root, blockSlot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) (ok bool) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.protoArray.ProcessBlock(parentRoot, blockRoot, blockSlot, justifiedEpoch, finalizedEpoch)
}

func (fc *ProtoForkChoice) InSubtree(anchor Root, root Root) (unknown bool, inSubtree bool) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.protoArray.InSubtree(anchor, root)
}

func (fc *ProtoForkChoice) Search(anchor NodeRef, parentRoot *Root, slot *Slot) (nonCanon []NodeRef, canon []NodeRef, err error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.protoArray.Search(anchor, parentRoot, slot)
}

func (fc *ProtoForkChoice) ClosestToSlot(anchor Root, slot Slot) (ref NodeRef, err error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.protoArray.ClosestToSlot(anchor, slot)
}

func (fc *ProtoForkChoice) CanonAtSlot(anchor Root, slot Slot, withBlock bool) (closest NodeRef, err error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.protoArray.CanonAtSlot(anchor, slot, withBlock)
}

func (fc *ProtoForkChoice) GetSlot(root Root) (slot Slot, ok bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.protoArray.GetSlot(root)
}

func (fc *ProtoForkChoice) FindHead(anchorRoot Root, anchorSlot Slot) (NodeRef, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	if err := fc.updateVotesMaybe(); err != nil {
		return NodeRef{}, err
	}
	return fc.protoArray.FindHead(anchorRoot, anchorSlot)
}

func (fc *ProtoForkChoice) Head() (NodeRef, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	if err := fc.updateVotesMaybe(); err != nil {
		return NodeRef{}, err
	}
	root := fc.justified.Root
	slot, _ := fc.spec.EpochStartSlot(fc.justified.Epoch)
	if fc.pin != nil {
		root = fc.pin.Root
		slot = fc.pin.Slot
	}
	return fc.protoArray.FindHead(root, slot)
}
