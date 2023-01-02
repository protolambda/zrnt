package operations

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"testing"

	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type WithdrawalsTestCase struct {
	test_util.BaseTransitionTest
	ExecutionPayload capella.ExecutionPayload
}

func (c *WithdrawalsTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	test_util.LoadSSZ(t, "execution_payload", c.Spec.Wrap(&c.ExecutionPayload), readPart)
}

func (c *WithdrawalsTestCase) Run() error {
	s, ok := c.Pre.(*capella.BeaconStateView)
	if !ok {
		return fmt.Errorf("unrecognized state type: %T", c.Pre)
	}
	return capella.ProcessWithdrawals(context.Background(), c.Spec, s, &c.ExecutionPayload)
}

func TestWithdrawals(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"capella"}, "operations", "withdrawals",
		func() test_util.TransitionTest { return new(WithdrawalsTestCase) })
}
