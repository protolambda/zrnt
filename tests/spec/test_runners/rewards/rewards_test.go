package epoch_processing

import (
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"strings"
	"testing"
)

type RewardsTest struct {
	Pre  *phase0.BeaconState

	Input *RewardsAndPenalties
	Output *RewardsAndPenalties

	// TODO refactor zrnt to split deltas
	ResDeltas *Deltas
}

func (c *RewardsTest) ExpectingFailure() bool {
	return false
}

func (c *RewardsTest) Load(t *testing.T, readPart test_util.TestPartReader) {
	pre := new(phase0.BeaconState)
	if test_util.LoadSSZ(t, "pre", pre, phase0.BeaconStateSSZ, readPart) {
		c.Pre = pre
	} else {
		t.Fatalf("failed to load pre state")
	}
	c.Input = &RewardsAndPenalties{}
	sourceDeltas := new(Deltas)
	if test_util.LoadSSZ(t, "source_deltas", sourceDeltas, DeltasSSZ, readPart) {
		c.Input.Source = sourceDeltas
	} else {
		t.Fatalf("failed to load source_deltas")
	}
	targetDeltas := new(Deltas)
	if test_util.LoadSSZ(t, "target_deltas", targetDeltas, DeltasSSZ, readPart) {
		c.Input.Target = targetDeltas
	} else {
		t.Fatalf("failed to load target_deltas")
	}
	headDeltas := new(Deltas)
	if test_util.LoadSSZ(t, "head_deltas", headDeltas, DeltasSSZ, readPart) {
		c.Input.Head = headDeltas
	} else {
		t.Fatalf("failed to load head_deltas")
	}
	inclusionDelayDeltas := new(Deltas)
	if test_util.LoadSSZ(t, "inclusion_delay_deltas", inclusionDelayDeltas, DeltasSSZ, readPart) {
		c.Input.InclusionDelay = inclusionDelayDeltas
	} else {
		t.Fatalf("failed to load inclusion_delay_deltas")
	}
	inactivityPenaltyDeltas := new(Deltas)
	if test_util.LoadSSZ(t, "inactivity_penalty_deltas", inactivityPenaltyDeltas, DeltasSSZ, readPart) {
		c.Input.Inactivity = inactivityPenaltyDeltas
	} else {
		t.Fatalf("failed to load inactivity_penalty_deltas")
	}
}

func (c *RewardsTest) Check(t *testing.T) {
	count := uint64(len(c.Input.Source.Rewards))
	diffDeltas := func(name string, computed *Deltas, expected *Deltas) {
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
	diffDeltas("inclusion delay", c.Output.InclusionDelay, c.Input.InclusionDelay)
	diffDeltas("inactivity", c.Output.Inactivity, c.Input.Inactivity)
}


func (c *RewardsTest) Run() error {
	state := phase0.NewFullFeaturedState(c.Pre)
	state.LoadPrecomputedData()
	c.Output = state.AttestationRewardsAndPenalties()
	return nil
}

func TestAllDeltas(t *testing.T) {
	test_util.RunTransitionTest(t, "rewards", "core", func() test_util.TransitionTest {
		return &RewardsTest{}
	})
}
