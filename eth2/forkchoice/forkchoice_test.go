package forkchoice

import "fmt"

type ForkChoiceTestTarget struct {
	// node -> canonical. Presence = pruneable node.
	// Test should fail if a node is pruned that is not in here.
	Pruneable map[NodeRef]bool
}

type Operation interface {
	Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error
}

type OpCanonicalChain struct {
	AnchorRoot Root
	AnchorSlot Slot
	// TODO: simplify to []Root ?
	Expected []ExtendedNodeRef
	Ok       bool
}

func (op *OpCanonicalChain) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
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
			return fmt.Errorf("entry %d differs: %s <> %s", ref, op.Expected[i])
		}
	}
	return nil
}

type OpClosestToSlot struct {
	Anchor  Root
	Slot    Slot
	Closest NodeRef
	Ok      bool
}

func (op *OpClosestToSlot) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
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
	Anchor    Root
	Slot      Slot
	WithBlock bool
	At        NodeRef
	Ok        bool
}

func (op *OpCanonAtSlot) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
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
	BlockRoot Root
	Slot      Slot
	Ok        bool
}

func (op *OpGetSlot) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
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
	AnchorRoot   Root
	AnchorSlot   Slot
	ExpectedHead NodeRef
	Ok           bool
}

func (op *OpFindHead) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
	head, err := fc.FindHead(op.AnchorRoot, op.AnchorSlot)
	if op.Ok && err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}
	if !op.Ok && err == nil {
		return fmt.Errorf("unexpected no error")
	}
	if head != op.ExpectedHead {
		return fmt.Errorf("different found head from anchor %s slot %d: %s <> %s",
			op.AnchorRoot, op.AnchorSlot, head, op.ExpectedHead)
	}
	return nil
}

type OpHead struct {
	ExpectedHead NodeRef
	Ok           bool
}

func (op *OpHead) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
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
	Root       Root
	OfRoot     Root
	Unknown    bool
	IsAncestor bool
}

func (op *OpIsAncestor) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
	unknown, isAncestor := fc.IsAncestor(op.Root, op.OfRoot)
	if unknown != op.Unknown || isAncestor != op.IsAncestor {
		return fmt.Errorf("expected different ancestor result: unknown %v <> %v, isAncestor %v <> %v",
			unknown, op.Unknown, isAncestor, op.IsAncestor)
	}
	return nil
}

type OpProcessSlot struct {
	Parent         Root
	Slot           Slot
	JustifiedEpoch Epoch
	FinalizedEpoch Epoch
}

func (op *OpProcessSlot) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
	fc.ProcessSlot(op.Parent, op.Slot, op.JustifiedEpoch, op.FinalizedEpoch)
	return nil
}

type OpProcessBlock struct {
	Parent         Root
	BlockRoot      Root
	BlockSlot      Slot
	JustifiedEpoch Epoch
	FinalizedEpoch Epoch
}

func (op *OpProcessBlock) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
	fc.ProcessBlock(op.Parent, op.BlockRoot, op.BlockSlot, op.JustifiedEpoch, op.FinalizedEpoch)
	return nil
}

type OpProcessAttestation struct {
	ValidatorIndex ValidatorIndex
	BlockRoot      Root
	HeadSlot       Slot
}

func (op *OpProcessAttestation) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
	fc.ProcessAttestation(op.ValidatorIndex, op.BlockRoot, op.HeadSlot)
	return nil
}

type OpPruneable struct {
	Pruneable NodeRef
	Canonical bool
}

func (op *OpPruneable) Apply(ft *ForkChoiceTestTarget, fc Forkchoice) error {
	ft.Pruneable[op.Pruneable] = op.Canonical
	return nil
}

type ForkChoiceTestInit struct {
	Finalized    Checkpoint
	Justified    Checkpoint
	AnchorRoot   Root
	AnchorSlot   Slot
	AnchorParent Root
}

type ForkChoiceTestDef struct {
	Init       ForkChoiceTestInit
	Operations []Operation
}

func (fd *ForkChoiceTestDef) Run(prepare func(init *ForkChoiceTestInit, ft *ForkChoiceTestTarget) Forkchoice) error {
	ft := &ForkChoiceTestTarget{
		Pruneable: make(map[NodeRef]bool),
	}
	fc := prepare(&fd.Init, ft)
	for i, op := range fd.Operations {
		if err := op.Apply(ft, fc); err != nil {
			return fmt.Errorf("test failed at step %d: %v", i, err)
		}
	}
	return nil
}
