package epoch_processing

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type stateFn func(spec *beacon.Spec, state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error

type EpochTest struct {
	test_util.BaseTransitionTest
	fn stateFn
}

func (c *EpochTest) Run() error {
	epc, err := c.Spec.NewEpochsContext(c.Pre)
	if err != nil {
		return err
	}
	process, err := c.Spec.PrepareEpochProcess(context.Background(), epc, c.Pre)
	if err != nil {
		return err
	}
	return c.fn(c.Spec, c.Pre, epc, process)
}

func NewEpochTest(fn stateFn) test_util.TransitionCaseMaker {
	return func() test_util.TransitionTest {
		return &EpochTest{fn: func(spec *beacon.Spec, state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return fn(spec, state, epc, process)
		}}
	}
}

func TestFinalUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "final_updates",
		NewEpochTest(func(spec *beacon.Spec, state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return spec.ProcessEpochFinalUpdates(context.Background(), epc, process, state)
		}))
}

func TestJustificationAndFinalization(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "justification_and_finalization",
		NewEpochTest(func(spec *beacon.Spec, state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return spec.ProcessEpochJustification(context.Background(), epc, process, state)
		}))
}

func TestRewardsPenalties(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "rewards_and_penalties",
		NewEpochTest(func(spec *beacon.Spec, state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return spec.ProcessEpochRewardsAndPenalties(context.Background(), epc, process, state)
		}))
}

func TestRegistryUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "registry_updates",
		NewEpochTest(func(spec *beacon.Spec, state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return spec.ProcessEpochRegistryUpdates(context.Background(), epc, process, state)
		}))
}

func TestSlashings(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "slashings",
		NewEpochTest(func(spec *beacon.Spec, state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return spec.ProcessEpochSlashings(context.Background(), epc, process, state)
		}))
}
