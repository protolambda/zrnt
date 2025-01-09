package operations

import (
	"context"
	"fmt"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type SyncAggregateTestCase struct {
	test_util.BaseTransitionTest
	SyncAggregate altair.SyncAggregate
}

func (c *SyncAggregateTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	test_util.LoadSSZ(t, "sync_aggregate", c.Spec.Wrap(&c.SyncAggregate), readPart)
}

func (c *SyncAggregateTestCase) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	s, ok := c.Pre.(altair.AltairLikeBeaconState)
	if !ok {
		return fmt.Errorf("unrecognized state type: %T", c.Pre)
	}
	return altair.ProcessSyncAggregate(context.Background(), c.Spec, epc, s, &c.SyncAggregate)
}

func TestSyncAggregate(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"altair", "bellatrix", "capella", "deneb"}, "operations", "sync_aggregate",
		func() test_util.TransitionTest { return new(SyncAggregateTestCase) })
}
