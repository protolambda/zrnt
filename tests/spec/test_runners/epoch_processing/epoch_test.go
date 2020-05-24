package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type stateFn func(state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error

type EpochTest struct {
	test_util.BaseTransitionTest
	fn stateFn
}

func (c *EpochTest) Run() error {
	epc, err := c.Pre.NewEpochsContext()
	if err != nil {
		return err
	}
	process, err := c.Pre.PrepareEpochProcess(epc)
	if err != nil {
		return err
	}
	return c.fn(c.Pre, epc, process)
}

func NewEpochTest(fn stateFn) test_util.TransitionCaseMaker {
	return func() test_util.TransitionTest {
		return &EpochTest{fn: func(state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return fn(state, epc, process)
		}}
	}
}

func TestFinalUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "final_updates",
		NewEpochTest(func(state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return state.ProcessEpochFinalUpdates(epc, process)
		}))
}

func TestJustificationAndFinalization(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "justification_and_finalization",
		NewEpochTest(func(state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return state.ProcessEpochJustification(epc, process)
		}))
}

func TestRewardsPenalties(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "rewards_and_penalties",
		NewEpochTest(func(state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return state.ProcessEpochRewardsAndPenalties(epc, process)
		}))
}

func TestRegistryUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "registry_updates",
		NewEpochTest(func(state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return state.ProcessEpochRegistryUpdates(epc, process)
		}))
}

func TestSlashings(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "slashings",
		NewEpochTest(func(state *beacon.BeaconStateView, epc *beacon.EpochsContext, process *beacon.EpochProcess) error {
			return state.ProcessEpochSlashings(epc, process)
		}))
}
