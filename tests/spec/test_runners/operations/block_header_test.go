package operations

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type BlockHeaderTestCase struct {
	test_util.BaseTransitionTest
	Block beacon.BeaconBlock
}

func (c *BlockHeaderTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	test_util.LoadSpecObj(t, "block", &c.Block, readPart)
}

func (c *BlockHeaderTestCase) Run() error {
	epc, err := c.Spec.NewEpochsContext(c.Pre)
	if err != nil {
		return err
	}
	return c.Spec.ProcessHeader(context.Background(), epc, c.Pre, &c.Block)
}

func TestBlockHeader(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "block_header",
		func() test_util.TransitionTest { return new(BlockHeaderTestCase) })
}
