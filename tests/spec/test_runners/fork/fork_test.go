package finality

import (
	"fmt"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"

	"gopkg.in/yaml.v3"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/bellatrix"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type ForkTestCase struct {
	Spec     *common.Spec
	PostFork test_util.ForkName
	Pre      common.BeaconState
	Post     common.BeaconState
}

type ForkMeta struct {
	Fork string `yaml:"fork"`
}

func (c *ForkTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.Spec = readPart.Spec()

	p := readPart.Part("meta.yaml")
	dec := yaml.NewDecoder(p)
	m := &ForkMeta{}
	test_util.Check(t, dec.Decode(&m))
	test_util.Check(t, p.Close())
	c.PostFork = test_util.ForkName(m.Fork)

	var preFork test_util.ForkName
	switch c.PostFork {
	case "altair":
		preFork = "phase0"
	case "bellatrix":
		preFork = "altair"
	case "capella":
		preFork = "bellatrix"
	case "deneb":
		preFork = "capella"
	default:
		t.Fatalf("unrecognized fork: %s", c.PostFork)
		return
	}

	if pre := test_util.LoadState(t, preFork, "pre", readPart); pre != nil {
		c.Pre = pre
	} else {
		t.Fatalf("failed to load pre state")
	}

	if post := test_util.LoadState(t, c.PostFork, "post", readPart); post != nil {
		c.Post = post
	}
}

func (c *ForkTestCase) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	switch c.PostFork {
	case "altair":
		out, err := altair.UpgradeToAltair(c.Spec, epc, c.Pre.(*phase0.BeaconStateView))
		if err != nil {
			return err
		}
		c.Pre = out
	case "bellatrix":
		out, err := bellatrix.UpgradeToBellatrix(c.Spec, epc, c.Pre.(*altair.BeaconStateView))
		if err != nil {
			return err
		}
		c.Pre = out
	case "capella":
		out, err := capella.UpgradeToCapella(c.Spec, epc, c.Pre.(*bellatrix.BeaconStateView))
		if err != nil {
			return err
		}
		c.Pre = out
	case "deneb":
		out, err := deneb.UpgradeToDeneb(c.Spec, epc, c.Pre.(*capella.BeaconStateView))
		if err != nil {
			return err
		}
		c.Pre = out
	default:
		return fmt.Errorf("unrecognized fork: %s", c.PostFork)
	}
	return nil
}

func (c *ForkTestCase) ExpectingFailure() bool {
	return false
}

func (c *ForkTestCase) Check(t *testing.T) {
	diff, err := test_util.CompareStates(c.Spec, c.Pre, c.Post)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Errorf("end result does not match expectation!\n%s", diff)
	}
}

func TestFork(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"altair", "bellatrix", "capella", "deneb"}, "fork", "fork",
		func() test_util.TransitionTest { return new(ForkTestCase) })
}
