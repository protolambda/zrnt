package fctest

import (
	"context"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/forkchoice"
)

type ForkChoiceTestTarget struct {
	// node -> canonical. Presence = pruneable node.
	// Test should fail if a node is pruned that is not in here.
	Pruneable map[forkchoice.NodeRef]bool
}

type Operation interface {
	Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error
}

type OpCanonicalChain struct {
	AnchorRoot forkchoice.Root
	AnchorSlot forkchoice.Slot
	// TODO: simplify to []Root ?
	Expected []forkchoice.ExtendedNodeRef
	Ok       bool
}

func (op *OpCanonicalChain) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	canon, err := fc.CanonicalChain(op.AnchorRoot, op.AnchorSlot)
	if op.Ok && err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}
	if !op.Ok && err == nil {
		return fmt.Errorf("unexpected no error")
	}
	if len(canon) != len(op.Expected) {
		return fmt.Errorf("expected different chain lengths: %d <> %d", len(canon), len(op.Expected))
	}
	for i, ref := range canon {
		if ref != op.Expected[i] {
			return fmt.Errorf("entry %d differs: %s <> %s", i, ref, op.Expected[i])
		}
	}
	return nil
}

type OpClosestToSlot struct {
	Anchor  forkchoice.Root
	Slot    forkchoice.Slot
	Closest forkchoice.NodeRef
	Ok      bool
}

func (op *OpClosestToSlot) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	closest, err := fc.ClosestToSlot(op.Anchor, op.Slot)
	if op.Ok && err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}
	if !op.Ok && err == nil {
		return fmt.Errorf("unexpected no error")
	}
	if closest != op.Closest {
		return fmt.Errorf("different closest node to %s slot %d: %s <> %s",
			op.Anchor, op.Slot, closest, op.Closest)
	}
	return nil
}

type OpCanonAtSlot struct {
	Anchor    forkchoice.Root
	Slot      forkchoice.Slot
	WithBlock bool
	At        forkchoice.NodeRef
	Ok        bool
}

func (op *OpCanonAtSlot) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	at, err := fc.CanonAtSlot(op.Anchor, op.Slot, op.WithBlock)
	if op.Ok && err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}
	if !op.Ok && err == nil {
		return fmt.Errorf("unexpected no error")
	}
	if at != op.At {
		return fmt.Errorf("different canon node from %s slot %d: %s <> %s",
			op.Anchor, op.Slot, at, op.At)
	}
	return nil
}

type OpGetSlot struct {
	BlockRoot forkchoice.Root
	Slot      forkchoice.Slot
	Ok        bool
}

func (op *OpGetSlot) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	slot, ok := fc.GetSlot(op.BlockRoot)
	if op.Ok && !ok {
		return fmt.Errorf("unexpected fail")
	}
	if !op.Ok && ok {
		return fmt.Errorf("unexpected no fail")
	}
	if slot != op.Slot {
		return fmt.Errorf("different slot for root %s: %d <> %d", op.BlockRoot, slot, op.Slot)
	}
	return nil
}

type OpFindHead struct {
	AnchorRoot   forkchoice.Root
	AnchorSlot   forkchoice.Slot
	ExpectedHead forkchoice.NodeRef
	Ok           bool
}

func (op *OpFindHead) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	head, err := fc.FindHead(op.AnchorRoot, op.AnchorSlot)
	if op.Ok && err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}
	if !op.Ok && err == nil {
		return fmt.Errorf("unexpected no error")
	}
	if head != op.ExpectedHead {
		return fmt.Errorf("different found head from pin %s slot %d: %s <> %s",
			op.AnchorRoot, op.AnchorSlot, head, op.ExpectedHead)
	}
	return nil
}

type OpHead struct {
	ExpectedHead forkchoice.NodeRef
	Ok           bool
}

func (op *OpHead) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	head, err := fc.Head()
	if op.Ok && err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}
	if !op.Ok && err == nil {
		return fmt.Errorf("unexpected no error")
	}
	if head != op.ExpectedHead {
		return fmt.Errorf("different head: %s <> %s", head, op.ExpectedHead)
	}
	return nil
}

type OpIsAncestor struct {
	Anchor    forkchoice.Root
	Root      forkchoice.Root
	Unknown   bool
	InSubtree bool
}

func (op *OpIsAncestor) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	unknown, inSubtree := fc.InSubtree(op.Anchor, op.Root)
	if unknown != op.Unknown || inSubtree != op.InSubtree {
		return fmt.Errorf("expected different in-subtree result: unknown %v <> %v, inSubtree %v <> %v",
			unknown, op.Unknown, inSubtree, op.InSubtree)
	}
	return nil
}

type OpProcessSlot struct {
	Parent         forkchoice.Root
	Slot           forkchoice.Slot
	JustifiedEpoch forkchoice.Epoch
	FinalizedEpoch forkchoice.Epoch
}

func (op *OpProcessSlot) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	fc.ProcessSlot(op.Parent, op.Slot, op.JustifiedEpoch, op.FinalizedEpoch)
	return nil
}

type OpProcessBlock struct {
	Parent         forkchoice.Root
	BlockRoot      forkchoice.Root
	BlockSlot      forkchoice.Slot
	JustifiedEpoch forkchoice.Epoch
	FinalizedEpoch forkchoice.Epoch
}

func (op *OpProcessBlock) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	fc.ProcessBlock(op.Parent, op.BlockRoot, op.BlockSlot, op.JustifiedEpoch, op.FinalizedEpoch)
	return nil
}

type OpProcessAttestation struct {
	ValidatorIndex forkchoice.ValidatorIndex
	BlockRoot      forkchoice.Root
	HeadSlot       forkchoice.Slot
	CanAdd         bool
}

func (op *OpProcessAttestation) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	res := fc.ProcessAttestation(op.ValidatorIndex, op.BlockRoot, op.HeadSlot)
	if res != op.CanAdd {
		return fmt.Errorf("processing attestation different result: canAdd %v <> %v", res, op.CanAdd)
	}
	return nil
}

type OpPruneable struct {
	Pruneable forkchoice.NodeRef
	Canonical bool
}

func (op *OpPruneable) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	ft.Pruneable[op.Pruneable] = op.Canonical
	return nil
}

type OpUpdateJustified struct {
	Trigger                forkchoice.Root
	Justified              forkchoice.Checkpoint
	Finalized              forkchoice.Checkpoint
	JustifiedStateBalances func() ([]forkchoice.Gwei, error)
	Ok                     bool
}

func (op *OpUpdateJustified) Apply(ft *ForkChoiceTestTarget, fc forkchoice.Forkchoice) error {
	err := fc.UpdateJustified(context.Background(), op.Trigger, op.Finalized, op.Justified, op.JustifiedStateBalances)
	if op.Ok && err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}
	if !op.Ok && err == nil {
		return fmt.Errorf("unexpected no error")
	}
	return nil
}

type ForkChoiceTestInit struct {
	Spec         *common.Spec
	Finalized    forkchoice.Checkpoint
	Justified    forkchoice.Checkpoint
	AnchorRoot   forkchoice.Root
	AnchorSlot   forkchoice.Slot
	AnchorParent forkchoice.Root
	Balances     []forkchoice.Gwei
}

type ForkChoiceTestDef struct {
	Init       ForkChoiceTestInit
	Operations []Operation
}

func (fd *ForkChoiceTestDef) Run(prepare func(init *ForkChoiceTestInit, ft *ForkChoiceTestTarget) (forkchoice.Forkchoice, error)) error {
	ft := &ForkChoiceTestTarget{
		Pruneable: make(map[forkchoice.NodeRef]bool),
	}
	fc, err := prepare(&fd.Init, ft)
	if err != nil {
		return fmt.Errorf("failed forkchoice preparation: %v", err)
	}
	for i, op := range fd.Operations {
		if err := op.Apply(ft, fc); err != nil {
			return fmt.Errorf("test failed at step %d: %v", i, err)
		}
	}
	return nil
}
