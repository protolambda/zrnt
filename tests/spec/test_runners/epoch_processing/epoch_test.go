package epoch_processing

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type stateFn func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error

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
	return c.fn(c.Spec, c.Fork, c.Pre, epc, flats)
}

func NewEpochTest(fn stateFn) test_util.TransitionCaseMaker {
	return func() test_util.TransitionTest {
		return &EpochTest{fn: func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return fn(spec, fork, state, epc, flats)
		}}
	}
}

func TestEffectiveBalanceUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "epoch_processing", "effective_balance_updates",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessEffectiveBalanceUpdates(context.Background(), spec, epc, flats, state)
		}))
}

func TestEth1DataReset(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "epoch_processing", "eth1_data_reset",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessEth1DataReset(context.Background(), spec, epc, state)
		}))
}

func TestHistoricalRootsUpdate(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"phase0", "altair", "bellatrix"}, "epoch_processing", "historical_roots_update",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessHistoricalRootsUpdate(context.Background(), spec, epc, state)
		}))
}

func TestHistoricalSummariesUpdate(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"capella"}, "epoch_processing", "historical_summaries_update",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return capella.ProcessHistoricalSummariesUpdate(context.Background(), spec, epc, state.(capella.HistoricalSummariesBeaconState))
		}))
}

func TestJustificationAndFinalization(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "epoch_processing", "justification_and_finalization",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			var just *phase0.JustificationStakeData
			if s, ok := state.(phase0.Phase0PendingAttestationsBeaconState); ok {
				attesterData, err := phase0.ComputeEpochAttesterData(context.Background(), spec, epc, flats, s)
				if err != nil {
					return err
				}
				just = &phase0.JustificationStakeData{
					CurrentEpoch:                  epc.CurrentEpoch.Epoch,
					TotalActiveStake:              epc.TotalActiveStake,
					PrevEpochUnslashedTargetStake: attesterData.PrevEpochUnslashedStake.TargetStake,
					CurrEpochUnslashedTargetStake: attesterData.CurrEpochUnslashedTargetStake,
				}
			} else if s, ok := state.(altair.AltairLikeBeaconState); ok {
				attesterData, err := altair.ComputeEpochAttesterData(context.Background(), spec, epc, flats, s)
				if err != nil {
					return err
				}
				just = &phase0.JustificationStakeData{
					CurrentEpoch:                  epc.CurrentEpoch.Epoch,
					TotalActiveStake:              epc.TotalActiveStake,
					PrevEpochUnslashedTargetStake: attesterData.PrevEpochUnslashedStake.TargetStake,
					CurrEpochUnslashedTargetStake: attesterData.CurrEpochUnslashedTargetStake,
				}
			} else {
				return fmt.Errorf("unrecognized state type: %T", state)
			}
			return phase0.ProcessEpochJustification(context.Background(), spec, just, state)
		}))
}

func TestParticipationRecordUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"phase0"}, "epoch_processing", "participation_record_updates",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			if s, ok := state.(phase0.Phase0PendingAttestationsBeaconState); ok {
				return phase0.ProcessParticipationRecordUpdates(context.Background(), spec, epc, s)
			} else {
				return fmt.Errorf("unrecognized state type: %T", state)
			}
		}))
}

func TestParticipationFlagUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"altair", "bellatrix", "capella", "deneb"}, "epoch_processing", "participation_flag_updates",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			if s, ok := state.(altair.AltairLikeBeaconState); ok {
				return altair.ProcessParticipationFlagUpdates(context.Background(), spec, s)
			} else {
				return fmt.Errorf("unrecognized state type: %T", state)
			}
		}))
}

func TestRandaoMixesReset(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "epoch_processing", "randao_mixes_reset",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessRandaoMixesReset(context.Background(), spec, epc, state)
		}))
}

func TestRegistryUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "epoch_processing", "registry_updates",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			switch fork {
			case "phase0", "altair", "bellatrix", "capella":
				return phase0.ProcessEpochRegistryUpdates(context.Background(), spec, epc, flats, state)
			case "deneb":
				return deneb.ProcessEpochRegistryUpdates(context.Background(), spec, epc, flats, state)
			default:
				return fmt.Errorf("unrecognized fork: %s", fork)
			}
		}))
}

func TestRewardsPenalties(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "epoch_processing", "rewards_and_penalties",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			if s, ok := state.(phase0.Phase0PendingAttestationsBeaconState); ok {
				attesterData, err := phase0.ComputeEpochAttesterData(context.Background(), spec, epc, flats, s)
				if err != nil {
					return err
				}
				return phase0.ProcessEpochRewardsAndPenalties(context.Background(), spec, epc, attesterData, s)
			} else if s, ok := state.(altair.AltairLikeBeaconState); ok {
				attesterData, err := altair.ComputeEpochAttesterData(context.Background(), spec, epc, flats, s)
				if err != nil {
					return err
				}
				return altair.ProcessEpochRewardsAndPenalties(context.Background(), spec, epc, attesterData, s)
			} else {
				return fmt.Errorf("unrecognized state type: %T", state)
			}
		}))
}

func TestSlashings(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "epoch_processing", "slashings",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessEpochSlashings(context.Background(), spec, epc, flats, state)
		}))
}

func TestSlashingsReset(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "epoch_processing", "slashings_reset",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			return phase0.ProcessSlashingsReset(context.Background(), spec, epc, state)
		}))
}

func TestSyncCommitteeUpdates(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"altair", "bellatrix", "capella", "deneb"}, "epoch_processing", "sync_committee_updates",
		NewEpochTest(func(spec *common.Spec, fork test_util.ForkName, state common.BeaconState, epc *common.EpochsContext, flats []common.FlatValidator) error {
			if s, ok := state.(common.SyncCommitteeBeaconState); ok {
				return altair.ProcessSyncCommitteeUpdates(context.Background(), spec, epc, s)
			} else {
				return fmt.Errorf("unrecognized state type: %T", state)
			}
		}))
}
