package epoch_processing

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type stateFn func(spec *common.Spec, state *phase0.BeaconStateView, epc *phase0.EpochsContext, process *phase0.EpochAttesterData) error

type EpochTest struct {
	test_util.BaseTransitionTest
	fn stateFn
}

func (c *EpochTest) Run() error {
	epc, err := phase0.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	process, err := phase0.ComputeEpochAttesterData(context.Background(), c.Spec, epc, c.Pre)
	if err != nil {
		return err
	}
	return c.fn(c.Spec, c.Pre, epc, process)
}

func NewEpochTest(fn stateFn) test_util.TransitionCaseMaker {
	return func() test_util.TransitionTest {
		return &EpochTest{fn: func(spec *common.Spec, state *phase0.BeaconStateView, epc *phase0.EpochsContext, process *phase0.EpochAttesterData) error {
			return fn(spec, state, epc, process)
		}}
	}
}

func TestFinalUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "final_updates",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *phase0.EpochsContext, process *phase0.EpochAttesterData) error {
			return phase0.ProcessEpochFinalUpdates(context.Background(), spec, epc, process, state)
		}))
}

func TestJustificationAndFinalization(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "justification_and_finalization",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *phase0.EpochsContext, process *phase0.EpochAttesterData) error {
			return phase0.ProcessEpochJustification(context.Background(), spec, epc, process, state)
		}))
}

func TestRewardsPenalties(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "rewards_and_penalties",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *phase0.EpochsContext, process *phase0.EpochAttesterData) error {
			return phase0.ProcessEpochRewardsAndPenalties(context.Background(), spec, epc, process, state)
		}))
}

func TestRegistryUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "registry_updates",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *phase0.EpochsContext, process *phase0.EpochAttesterData) error {
			return phase0.ProcessEpochRegistryUpdates(context.Background(), spec, epc, process, state)
		}))
}

func TestSlashings(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "slashings",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *phase0.EpochsContext, process *phase0.EpochAttesterData) error {
			return phase0.ProcessEpochSlashings(context.Background(), spec, epc, process, state)
		}))
}
