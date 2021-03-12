package operations

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type BlockHeaderTestCase struct {
	test_util.BaseTransitionTest
	Block phase0.BeaconBlock
}

func (c *BlockHeaderTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	test_util.LoadSpecObj(t, "block", &c.Block, readPart)
}

func (c *BlockHeaderTestCase) Run() error {
	epc, err := phase0.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	return phase0.ProcessHeader(context.Background(), c.Spec, epc, c.Pre, &c.Block)
}

func TestBlockHeader(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "block_header",
		func() test_util.TransitionTest { return new(BlockHeaderTestCase) })
}
