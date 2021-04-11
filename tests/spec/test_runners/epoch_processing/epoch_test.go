package epoch_processing

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type stateFn func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error

type EpochTest struct {
	test_util.BaseTransitionTest
	fn stateFn
}

func (c *EpochTest) Run() error {
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
	return c.fn(c.Spec, c.Pre, epc, flats)
}

func NewEpochTest(fn stateFn) test_util.TransitionCaseMaker {
	return func() test_util.TransitionTest {
		return &EpochTest{fn: func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return fn(spec, state, epc, flats)
		}}
	}
}

func TestEffectiveBalanceUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "effective_balance_updates",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessEffectiveBalanceUpdates(context.Background(), spec, epc, flats, state)
		}))
}

func TestEth1DataReset(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "eth1_data_reset",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessEth1DataReset(context.Background(), spec, epc, state)
		}))
}

func TestHistoricalRootsUpdate(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "historical_roots_update",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessHistoricalRootsUpdate(context.Background(), spec, epc, state)
		}))
}

func TestJustificationAndFinalization(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "justification_and_finalization",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			attesterData, err := phase0.ComputeEpochAttesterData(context.Background(), spec, epc, flats, state)
			if err != nil {
				return err
			}
			just := phase0.JustificationStakeData{
				CurrentEpoch:                  epc.CurrentEpoch.Epoch,
				TotalActiveStake:              epc.TotalActiveStake,
				PrevEpochUnslashedTargetStake: attesterData.PrevEpochUnslashedStake.TargetStake,
				CurrEpochUnslashedTargetStake: attesterData.CurrEpochUnslashedTargetStake,
			}
			return phase0.ProcessEpochJustification(context.Background(), spec, &just, state)
		}))
}

func TestParticipationRecordUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "participation_record_updates",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessParticipationRecordUpdates(context.Background(), spec, epc, state)
		}))
}

func TestRandaoMixesReset(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "randao_mixes_reset",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessRandaoMixesReset(context.Background(), spec, epc, state)
		}))
}

func TestRegistryUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "registry_updates",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessEpochRegistryUpdates(context.Background(), spec, epc, flats, state)
		}))
}

func TestRewardsPenalties(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "rewards_and_penalties",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			attesterData, err := phase0.ComputeEpochAttesterData(context.Background(), spec, epc, flats, state)
			if err != nil {
				return err
			}
			return phase0.ProcessEpochRewardsAndPenalties(context.Background(), spec, epc, attesterData, state)
		}))
}

func TestSlashings(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "slashings",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessEpochSlashings(context.Background(), spec, epc, flats, state)
		}))
}

func TestSlashingsReset(t *testing.T) {
	test_util.RunTransitionTest(t, "epoch_processing", "slashings_reset",
		NewEpochTest(func(spec *common.Spec, state *phase0.BeaconStateView, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessSlashingsReset(context.Background(), spec, epc, state)
		}))
}
