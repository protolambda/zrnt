package altair

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

type EpochStakeSummary struct {
	SourceStake common.Gwei
	TargetStake common.Gwei
	HeadStake   common.Gwei
}

type EpochAttesterData struct {
	Flats []common.FlatValidator

	PrevEpochUnslashedStake       EpochStakeSummary
	CurrEpochUnslashedTargetStake common.Gwei
}

func ComputeEpochAttesterData(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, flats []common.FlatValidator, state *BeaconStateView) (out *EpochAttesterData, err error) {
	out = &EpochAttesterData{
		Flats: flats,
		PrevEpochUnslashedStake: EpochStakeSummary{
			SourceStake: 0,
			TargetStake: 0,
			HeadStake:   0,
		},
		CurrEpochUnslashedTargetStake: 0,
	}
	prevEpochParticipationView, err := state.PreviousEpochParticipation()
	if err != nil {
		return nil, err
	}
	prevEpochParticipation, err := prevEpochParticipationView.Raw()
	if err != nil {
		return nil, err
	}
	currEpochParticipationView, err := state.CurrentEpochParticipation()
	if err != nil {
		return nil, err
	}
	currEpochParticipation, err := currEpochParticipationView.Raw()
	if err != nil {
		return nil, err
	}
	for _, vi := range epc.PreviousEpoch.ActiveIndices {
		if flats[vi].Slashed {
			continue
		}
		effBal := flats[vi].EffectiveBalance
		prevFlag := prevEpochParticipation[vi]
		if prevFlag&TIMELY_SOURCE_FLAG != 0 {
			out.PrevEpochUnslashedStake.SourceStake += effBal
		}
		if prevFlag&TIMELY_TARGET_FLAG != 0 {
			out.PrevEpochUnslashedStake.TargetStake += effBal
		}
		if prevFlag&TIMELY_HEAD_FLAG != 0 {
			out.PrevEpochUnslashedStake.HeadStake += effBal
		}
		if currEpochParticipation[vi]&TIMELY_TARGET_FLAG != 0 {
			out.CurrEpochUnslashedTargetStake += effBal
		}
	}
	if out.PrevEpochUnslashedStake.SourceStake < spec.EFFECTIVE_BALANCE_INCREMENT {
		out.PrevEpochUnslashedStake.SourceStake = spec.EFFECTIVE_BALANCE_INCREMENT
	}
	if out.PrevEpochUnslashedStake.TargetStake < spec.EFFECTIVE_BALANCE_INCREMENT {
		out.PrevEpochUnslashedStake.TargetStake = spec.EFFECTIVE_BALANCE_INCREMENT
	}
	if out.PrevEpochUnslashedStake.HeadStake < spec.EFFECTIVE_BALANCE_INCREMENT {
		out.PrevEpochUnslashedStake.HeadStake = spec.EFFECTIVE_BALANCE_INCREMENT
	}
	if out.CurrEpochUnslashedTargetStake < spec.EFFECTIVE_BALANCE_INCREMENT {
		out.CurrEpochUnslashedTargetStake = spec.EFFECTIVE_BALANCE_INCREMENT
	}

	return out, nil
}
