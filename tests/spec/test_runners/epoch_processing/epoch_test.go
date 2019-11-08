package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type stateFn func(state *phase0.FullFeaturedState)

type EpochTest struct {
	test_util.BaseTransitionTest
	fn stateFn
}

func (c *EpochTest) Run() error {
	c.fn(c.Prepare())
	return nil
}

func NewEpochTest(fn stateFn) test_util.TransitionCaseMaker {
	return func() test_util.TransitionTest {
		return &EpochTest{fn: func(state *phase0.FullFeaturedState) {
			fn(state)
		}}
	}
}

func TestFinalUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "final_updates",
		NewEpochTest(func(state *phase0.FullFeaturedState) {
			state.ProcessEpochFinalUpdates()
		}))
}

func TestJustificationAndFinalization(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "justification_and_finalization",
		NewEpochTest(func(state *phase0.FullFeaturedState) {
			state.ProcessEpochJustification()
		}))
}

func TestRegistryUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "registry_updates",
		NewEpochTest(func(state *phase0.FullFeaturedState) {
			state.ProcessEpochRegistryUpdates()
		}))
}

func TestSlashings(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "slashings",
		NewEpochTest(func(state *phase0.FullFeaturedState) {
			state.ProcessEpochSlashings()
		}))
}
