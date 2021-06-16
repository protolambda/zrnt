package sanity

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/merge"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v3"
	"testing"
)

type BlocksTestCase struct {
	test_util.BaseTransitionTest
	Blocks []*common.BeaconBlockEnvelope
}

type BlocksCountMeta struct {
	BlocksCount uint64 `yaml:"blocks_count"`
}

func (c *BlocksTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	p := readPart.Part("meta.yaml")
	dec := yaml.NewDecoder(p)
	m := &BlocksCountMeta{}
	test_util.Check(t, dec.Decode(&m))
	test_util.Check(t, p.Close())
	valRoot, err := c.Pre.GenesisValidatorsRoot()
	if err != nil {
		t.Fatalf("failed to get pre-state genesis validators root: %v", err)
	}
	loadBlock := func(i uint64) *common.BeaconBlockEnvelope {
		switch forkName {
		case "phase0":
			dst := new(phase0.SignedBeaconBlock)
			test_util.LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.GENESIS_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		case "altair":
			dst := new(altair.SignedBeaconBlock)
			test_util.LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.ALTAIR_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		case "merge":
			dst := new(merge.SignedBeaconBlock)
			test_util.LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.MERGE_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		default:
			t.Fatalf("unrecognized fork name: %s", forkName)
			return nil
		}
	}
	for i := uint64(0); i < m.BlocksCount; i++ {
		c.Blocks = append(c.Blocks, loadBlock(i))
	}
}

func (c *BlocksTestCase) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	state := &beacon.StandardUpgradeableBeaconState{BeaconState: c.Pre}
	defer func() {
		c.Pre = state.BeaconState
	}()
	for _, b := range c.Blocks {
		if err := common.StateTransition(context.Background(), c.Spec, epc, state, b, true); err != nil {
			return err
		}
	}
	return nil
}

func TestBlocks(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "sanity", "blocks",
		func() test_util.TransitionTest { return new(BlocksTestCase) })
}
