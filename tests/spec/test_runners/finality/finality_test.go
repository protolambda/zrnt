package finality

import (
	"context"
	"fmt"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"

	"gopkg.in/yaml.v3"

	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/bellatrix"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type FinalityTestCase struct {
	test_util.BaseTransitionTest
	Blocks []*common.BeaconBlockEnvelope
}

type BlocksCountMeta struct {
	BlocksCount uint64 `yaml:"blocks_count"`
}

func (c *FinalityTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	p := readPart.Part("meta.yaml")
	dec := yaml.NewDecoder(p)
	m := &BlocksCountMeta{}
	test_util.Check(t, dec.Decode(&m))
	test_util.Check(t, p.Close())
	valRoot, err := c.Pre.GenesisValidatorsRoot()
	test_util.Check(t, err)

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
		case "bellatrix":
			dst := new(bellatrix.SignedBeaconBlock)
			test_util.LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.BELLATRIX_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		case "capella":
			dst := new(capella.SignedBeaconBlock)
			test_util.LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.CAPELLA_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		case "deneb":
			dst := new(deneb.SignedBeaconBlock)
			test_util.LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.DENEB_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		default:
			t.Fatal(fmt.Errorf("unrecognized fork name: %s", forkName))
			return nil
		}
	}

	for i := uint64(0); i < m.BlocksCount; i++ {
		c.Blocks = append(c.Blocks, loadBlock(i))
	}
}

func (c *FinalityTestCase) Run() error {
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
	test_util.RunTransitionTest(t, test_util.AllForks, "finality", "finality",
		func() test_util.TransitionTest { return new(FinalityTestCase) })
}
