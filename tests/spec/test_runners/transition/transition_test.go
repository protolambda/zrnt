package transition

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

type TransitionTestCase struct {
	test_util.BaseTransitionTest
	Blocks []*common.BeaconBlockEnvelope
}

type TransitionMeta struct {
	PostFork    string  `yaml:"post_fork"`
	ForkEpoch   uint64  `yaml:"fork_epoch"`
	ForkBlock   *uint64 `yaml:"fork_block"`
	BlocksCount uint64  `yaml:"blocks_count"`
}

func (c *TransitionTestCase) Load(t *testing.T, testFork test_util.ForkName, readPart test_util.TestPartReader) {
	// copy spec before modifying
	specCopy := *readPart.Spec()
	c.Spec = &specCopy

	p := readPart.Part("meta.yaml")
	dec := yaml.NewDecoder(p)
	var m TransitionMeta
	test_util.Check(t, dec.Decode(&m))
	test_util.Check(t, p.Close())

	var preForkName test_util.ForkName
	switch testFork {
	case "phase0":
		t.Fatal("cannot transition into phase0")
	case "altair":
		preForkName = "phase0"
		c.Spec.ALTAIR_FORK_EPOCH = common.Epoch(m.ForkEpoch)
	case "bellatrix":
		preForkName = "altair"
		c.Spec.BELLATRIX_FORK_EPOCH = common.Epoch(m.ForkEpoch)
	case "capella":
		preForkName = "bellatrix"
		c.Spec.CAPELLA_FORK_EPOCH = common.Epoch(m.ForkEpoch)
	case "deneb":
		preForkName = "capella"
		c.Spec.DENEB_FORK_EPOCH = common.Epoch(m.ForkEpoch)
	default:
		t.Fatalf("unsupported fork %s", testFork)
	}

	c.Fork = preForkName
	if pre := test_util.LoadState(t, preForkName, "pre", readPart); pre != nil {
		c.Pre = pre
	} else {
		t.Fatalf("failed to load pre state")
	}

	valRoot, err := c.Pre.GenesisValidatorsRoot()
	if err != nil {
		t.Fatalf("failed to get pre-state genesis validators root: %v", err)
	}

	if post := test_util.LoadState(t, test_util.ForkName(m.PostFork), "post", readPart); post != nil {
		c.Post = post
	} else {
		t.Fatalf("failed to load post state")
	}

	loadBlock := func(i uint64) *common.BeaconBlockEnvelope {
		forkName := preForkName
		if m.ForkBlock == nil || i > *m.ForkBlock {
			forkName = test_util.ForkName(m.PostFork)
		}
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
			t.Fatalf("unrecognized fork name: %s", forkName)
			return nil
		}
	}
	for i := uint64(0); i < m.BlocksCount; i++ {
		c.Blocks = append(c.Blocks, loadBlock(i))
	}
}

func (c *TransitionTestCase) Run() error {
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

func TestTransition(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"altair", "bellatrix", "capella", "deneb"}, "transition", "core",
		func() test_util.TransitionTest { return new(TransitionTestCase) })
}
