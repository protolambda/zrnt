package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon/header"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type BlockHeaderTestCase struct {
	test_util.BaseTransitionTest
	BlockHeader *header.BeaconBlockHeader
}

func (c *BlockHeaderTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	b := phase0.BeaconBlock{}
	test_util.LoadSSZ(t, "block", &b, phase0.BeaconBlockSSZ, readPart)
	c.BlockHeader = b.Header()
}

func (c *BlockHeaderTestCase) Run() error {
	state := c.Prepare()
	return state.ProcessHeader(c.BlockHeader)
}

func TestBlockHeader(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "block_header",
		func() test_util.TransitionTest { return new(BlockHeaderTestCase) })
}
