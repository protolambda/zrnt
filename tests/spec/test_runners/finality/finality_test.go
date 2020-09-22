package finality

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v2"
	"testing"
)

type FinalityTestCase struct {
	test_util.BaseTransitionTest
	Blocks []*beacon.SignedBeaconBlock
}

type BlocksCountMeta struct {
	BlocksCount uint64 `yaml:"blocks_count"`
}

func (c *FinalityTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	p := readPart.Part("meta.yaml")
	dec := yaml.NewDecoder(p)
	m := &BlocksCountMeta{}
	test_util.Check(t, dec.Decode(&m))
	test_util.Check(t, p.Close())
	loadBlock := func(i uint64) *beacon.SignedBeaconBlock {
		dst := new(beacon.SignedBeaconBlock)
		test_util.LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
		return dst
	}
	for i := uint64(0); i < m.BlocksCount; i++ {
		c.Blocks = append(c.Blocks, loadBlock(i))
	}
}

func (c *FinalityTestCase) Run() error {
	epc, err := c.Spec.NewEpochsContext(c.Pre)
	if err != nil {
		return err
	}
	state := c.Pre
	for _, b := range c.Blocks {
		if err := c.Spec.StateTransition(context.Background(), epc, state, b, true); err != nil {
			return err
		}
	}
	return nil
}

func TestBlocks(t *testing.T) {
	test_util.RunTransitionTest(t, "finality", "finality",
		func() test_util.TransitionTest { return new(FinalityTestCase) })
}
