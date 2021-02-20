package forkchoice

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"sync"
)

type ProtoForkChoice struct {
	mu         sync.RWMutex
	protoArray ForkchoiceGraph
	voteStore  VoteStore

	balances []Gwei
	// The block root and slot to start forkchoice from.
	// At genesis, explicitly set to genesis, since there is no "justified" root yet.
	// Later updated with justified state.
	// May be modified to pin a sub-tree for soft-fork purposes.
	anchor    NodeRef
	justified Checkpoint
	finalized Checkpoint
	spec      *beacon.Spec
}

var _ Forkchoice = (*ProtoForkChoice)(nil)

func NewForkChoice(spec *beacon.Spec, finalized Checkpoint, justified Checkpoint,
	anchorRoot Root, anchorSlot Slot, anchorParent Root, graph ForkchoiceGraph, votes VoteStore) *ProtoForkChoice {
	fc := &ProtoForkChoice{
		protoArray: graph,
		voteStore:  votes,
		balances:   nil,
		justified:  justified,
		finalized:  finalized,
		spec:       spec,
	}
	fc.ProcessBlock(anchorParent, anchorRoot, anchorSlot, justified.Epoch, finalized.Epoch)
	fc.anchor = NodeRef{Root: anchorRoot, Slot: anchorSlot}
	return fc
}

func (fc *ProtoForkChoice) PinAnchor(root Root, slot Slot) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
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
func (fc *ProtoForkChoice) UpdateJustified(ctx context.Context, justified Checkpoint, finalized Checkpoint,
	justifiedStateBalances func() ([]Gwei, error)) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	// Old/same data? Ignore the change.
	if fc.justified.Epoch >= justified.Epoch && fc.finalized.Epoch >= finalized.Epoch {
		return nil
	}
	justifiedSlot, err := fc.spec.EpochStartSlot(justified.Epoch)
	if err != nil {
		return fmt.Errorf("bad justified epoch: %d", justified.Epoch)
	}
	if justified.Epoch < finalized.Epoch {
		return fmt.Errorf("justified epoch %d lower than finalized epoch %d", justified.Epoch, finalized.Epoch)
	}

	// check if new finalized checkpoint is valid
	if fc.finalized != finalized {
		if unknown, isAncestor := fc.IsAncestor(finalized.Root, fc.finalized.Root); unknown {
			return fmt.Errorf("unknown finalized checkpoint: %s", finalized)
		} else if !isAncestor || fc.finalized.Epoch > finalized.Epoch {
			return fmt.Errorf("new finalized checkpoint %s is outside of finalized subtree: %s",
				finalized, fc.finalized)
		}
	}
	if fc.justified != justified {
		if unknown, isAncestor := fc.IsAncestor(justified.Root, fc.finalized.Root); unknown {
			return fmt.Errorf("unknown justified checkpoint: %s", justified)
		} else if !isAncestor || fc.finalized.Epoch > justified.Epoch {
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
	prevFinalized := fc.finalized
	fc.finalized = finalized
	fc.anchor = NodeRef{Root: justified.Root, Slot: justifiedSlot}

	// prune if we finalized something
	if prevFinalized != finalized {
		finSlot, _ := fc.spec.EpochStartSlot(finalized.Epoch)
		if err := fc.protoArray.OnPrune(ctx, finalized.Root, finSlot); err != nil {
			return err
		}
	}
	return nil
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

func (fc *ProtoForkChoice) ProcessAttestation(index ValidatorIndex, blockRoot Root, headSlot Slot) {
	fc.mu.Lock()
	fc.mu.Unlock()
	fc.voteStore.ProcessAttestation(index, blockRoot, headSlot)
}

func (fc *ProtoForkChoice) CanonicalChain(anchorRoot Root, anchorSlot Slot) ([]ExtendedNodeRef, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.CanonicalChain(anchorRoot, anchorSlot)
}

func (fc *ProtoForkChoice) ProcessSlot(parentRoot Root, slot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.protoArray.ProcessSlot(parentRoot, slot, justifiedEpoch, finalizedEpoch)
}

func (fc *ProtoForkChoice) ProcessBlock(parentRoot Root, blockRoot Root, blockSlot Slot, justifiedEpoch Epoch, finalizedEpoch Epoch) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.protoArray.ProcessBlock(parentRoot, blockRoot, blockSlot, justifiedEpoch, finalizedEpoch)
}

func (fc *ProtoForkChoice) IsAncestor(root Root, ofRoot Root) (unknown bool, isAncestor bool) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.protoArray.IsAncestor(root, ofRoot)
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
	return fc.protoArray.FindHead(anchorRoot, anchorSlot)
}

func (fc *ProtoForkChoice) Head() (NodeRef, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.protoArray.FindHead(fc.anchor.Root, fc.anchor.Slot)
}
