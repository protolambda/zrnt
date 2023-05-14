package operations

import (
	"context"
	"fmt"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"

	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type WithdrawalsTestCase struct {
	test_util.BaseTransitionTest
	ExecutionPayload capella.ExecutionPayloadWithWithdrawals
}

func (c *WithdrawalsTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	switch forkName {
	case "capella":
		var payload capella.ExecutionPayload
		test_util.LoadSSZ(t, "execution_payload", c.Spec.Wrap(&payload), readPart)
		c.ExecutionPayload = &payload
	case "deneb":
		var payload deneb.ExecutionPayload
		test_util.LoadSSZ(t, "execution_payload", c.Spec.Wrap(&payload), readPart)
		c.ExecutionPayload = &payload
	default:
		t.Fatalf("unrecognized fork: %s", forkName)
	}
}

func (c *WithdrawalsTestCase) Run() error {
	s, ok := c.Pre.(capella.BeaconStateWithWithdrawals)
	if !ok {
		return fmt.Errorf("unrecognized state type: %T", c.Pre)
	}
	return capella.ProcessWithdrawals(context.Background(), c.Spec, s, c.ExecutionPayload)
}

func TestWithdrawals(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"capella", "deneb"}, "operations", "withdrawals",
		func() test_util.TransitionTest { return new(WithdrawalsTestCase) })
}
