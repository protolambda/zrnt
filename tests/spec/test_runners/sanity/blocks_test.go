package sanity

import (
	"fmt"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v2"
	"testing"
)

type BlocksTestCase struct {
	test_util.BaseTransitionTest
	Blocks []*phase0.BeaconBlock
}

type BlocksCountMeta struct {
	BlocksCount uint64 `yaml:"blocks_count"`
}

func (c *BlocksTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	p := readPart("meta.yaml")
	dec := yaml.NewDecoder(p)
	m := &BlocksCountMeta{}
	test_util.Check(t, dec.Decode(&m))
	test_util.Check(t, p.Close())
	loadBlock := func(i uint64) *phase0.BeaconBlock {
		dst := new(phase0.BeaconBlock)
		test_util.LoadSSZ(t, fmt.Sprintf("blocks_%d", i), dst, phase0.BeaconBlockSSZ, readPart)
		return dst
	}
	for i := uint64(0); i < m.BlocksCount; i++ {
		c.Blocks = append(c.Blocks, loadBlock(i))
	}
}

func (c *BlocksTestCase) Run() error {
	state := c.Prepare()
	for _, b := range c.Blocks {
		blockProc := &phase0.BlockProcessFeature{Block: b, Meta: state}
		if err := state.StateTransition(blockProc, true); err != nil {
			return err
		}
	}
	return nil
}

func TestBlocks(t *testing.T) {
	test_util.RunTransitionTest(t, "sanity", "blocks",
		func() test_util.TransitionTest { return new(BlocksTestCase) })
}
