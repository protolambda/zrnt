package proto

import (
	"context"
	"fmt"
	"testing"

	"github.com/protolambda/zrnt/eth2/forkchoice"
	"github.com/protolambda/zrnt/eth2/forkchoice/internal/fctest"
)

func TestProtoArray(t *testing.T) {
	lhtest := fctest.LighthouseTestDef()
	err := lhtest.Run(func(init *fctest.ForkChoiceTestInit, ft *fctest.ForkChoiceTestTarget) (forkchoice.Forkchoice, error) {
		return NewProtoForkChoice(init.Spec, init.Finalized, init.Justified, init.AnchorRoot, init.AnchorSlot, init.AnchorParent, init.Balances,
			NodeSinkFn(func(ctx context.Context, ref forkchoice.NodeRef, canonical bool) error {
				// whenever something is pruned, check if it was allowed to be pruned,
				// and if it's marked as canonical correctly.
				expectedCanonical, ok := ft.Pruneable[ref]
				if !ok {
					return fmt.Errorf("unexpected pruning of node %s", ref)
				}
				if canonical != expectedCanonical {
					return fmt.Errorf("bad pruning, pruned as canonical=%v, but expected %v", canonical, expectedCanonical)
				}
				return nil
			}))
	})
	if err != nil {
		t.Error(err)
	}
}
