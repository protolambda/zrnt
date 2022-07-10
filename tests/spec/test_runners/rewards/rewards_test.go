package epoch_processing

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type RewardsTest struct {
	Spec *common.Spec
	Pre  common.BeaconState

	Input struct {
		Source *common.Deltas
		Target *common.Deltas
		Head   *common.Deltas
		// missing in Altair+
		InclusionDelay *common.Deltas
		Inactivity     *common.Deltas
	}
	Output struct {
		Source *common.Deltas
		Target *common.Deltas
		Head   *common.Deltas
		// missing in Altair+
		InclusionDelay *common.Deltas
		Inactivity     *common.Deltas
	}
}

func (c *RewardsTest) ExpectingFailure() bool {
	return false
}

func (c *RewardsTest) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.Spec = readPart.Spec()

	c.Pre = test_util.LoadState(t, forkName, "pre", readPart)

	sourceDeltas := new(common.Deltas)
	if test_util.LoadSpecObj(t, "source_deltas", sourceDeltas, readPart) {
		c.Input.Source = sourceDeltas
	} else {
		t.Fatalf("failed to load source_deltas")
	}
	targetDeltas := new(common.Deltas)
	if test_util.LoadSpecObj(t, "target_deltas", targetDeltas, readPart) {
		c.Input.Target = targetDeltas
	} else {
		t.Fatalf("failed to load target_deltas")
	}
	headDeltas := new(common.Deltas)
	if test_util.LoadSpecObj(t, "head_deltas", headDeltas, readPart) {
		c.Input.Head = headDeltas
	} else {
		t.Fatalf("failed to load head_deltas")
	}
	if forkName == "phase0" {
		inclusionDelayDeltas := new(common.Deltas)
		if test_util.LoadSpecObj(t, "inclusion_delay_deltas", inclusionDelayDeltas, readPart) {
			c.Input.InclusionDelay = inclusionDelayDeltas
		} else {
			t.Fatalf("failed to load inclusion_delay_deltas")
		}
	}
	inactivityPenaltyDeltas := new(common.Deltas)
	if test_util.LoadSpecObj(t, "inactivity_penalty_deltas", inactivityPenaltyDeltas, readPart) {
		c.Input.Inactivity = inactivityPenaltyDeltas
	} else {
		t.Fatalf("failed to load inactivity_penalty_deltas")
	}
}

func (c *RewardsTest) Check(t *testing.T) {
	count := uint64(len(c.Input.Source.Rewards))
	diffDeltas := func(name string, computed *common.Deltas, expected *common.Deltas) {
		t.Run(name, func(t *testing.T) {
			var failed bool
			var buf strings.Builder
			for i := uint64(0); i < count; i++ {
				if computed.Rewards[i] != expected.Rewards[i] {
					buf.WriteString(fmt.Sprintf("(%s) invalid reward: i: %d, expected: %d, got: %d\n",
						name, i, expected.Rewards[i], computed.Rewards[i]))
					failed = true
				}
				if computed.Penalties[i] != expected.Penalties[i] {
					buf.WriteString(fmt.Sprintf("(%s) invalid penalty: i: %d, expected: %d, got: %d\n",
						name, i, expected.Penalties[i], computed.Penalties[i]))
					failed = true
				}
			}
			if failed {
				t.Error("rewards error:\n" + buf.String())
			}
		})
	}
	diffDeltas("source", c.Output.Source, c.Input.Source)
	diffDeltas("target", c.Output.Target, c.Input.Target)
	diffDeltas("head", c.Output.Head, c.Input.Head)
	if c.Input.InclusionDelay != nil {
		diffDeltas("inclusion delay", c.Output.InclusionDelay, c.Input.InclusionDelay)
	}
	diffDeltas("inactivity", c.Output.Inactivity, c.Input.Inactivity)
}

func (c *RewardsTest) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	vals, err := c.Pre.Validators()
	if err != nil {
		return err
	}
	flats, err := common.FlattenValidators(vals)
	if err != nil {
		return err
	}

	if s, ok := c.Pre.(phase0.Phase0PendingAttestationsBeaconState); ok {
		attesterData, err := phase0.ComputeEpochAttesterData(context.Background(), c.Spec, epc, flats, s)
		if err != nil {
			return err
		}
		deltas, err := phase0.AttestationRewardsAndPenalties(context.Background(), c.Spec, epc, attesterData, s)
		if err != nil {
			return err
		}
		c.Output.Source = deltas.Source
		c.Output.Target = deltas.Target
		c.Output.Head = deltas.Head
		c.Output.InclusionDelay = deltas.InclusionDelay
		c.Output.Inactivity = deltas.Inactivity
	} else if s, ok := c.Pre.(altair.AltairLikeBeaconState); ok {
		attesterData, err := altair.ComputeEpochAttesterData(context.Background(), c.Spec, epc, flats, s)
		if err != nil {
			return err
		}
		deltas, err := altair.AttestationRewardsAndPenalties(context.Background(), c.Spec, epc, attesterData, s)
		if err != nil {
			return err
		}
		c.Output.Source = deltas.Source
		c.Output.Target = deltas.Target
		c.Output.Head = deltas.Head
		// no inclusion delay penalties during epoch in altair+, part of the block processing instead.
		c.Output.Inactivity = deltas.Inactivity
	} else {
		return fmt.Errorf("unrecognized state type: %T", c.Pre)
	}
	return err
}

func TestBasic(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "rewards", "basic", func() test_util.TransitionTest {
		return &RewardsTest{}
	})
}

func TestLeak(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "rewards", "leak", func() test_util.TransitionTest {
		return &RewardsTest{}
	})
}

func TestRandom(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "rewards", "random", func() test_util.TransitionTest {
		return &RewardsTest{}
	})
}
