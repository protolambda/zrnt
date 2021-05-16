package altair

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func ComputeFlagDeltas(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, attesterData *EpochAttesterData,
	flag ParticipationFlags, weight common.Gwei, isInactivityLeak bool) (*common.Deltas, error) {

	valCount := uint64(len(attesterData.Flats))
	out := common.NewDeltas(valCount)

	unslashedParticipatingTotalBalance := common.Gwei(0)
	for _, vi := range epc.PreviousEpoch.ActiveIndices {
		if !attesterData.Flats[vi].Slashed && (attesterData.PrevParticipation[vi]&flag != 0) {
			unslashedParticipatingTotalBalance += attesterData.Flats[vi].EffectiveBalance
		}
	}
	// get_total_balance makes it 1 increment minimum
	if unslashedParticipatingTotalBalance < spec.EFFECTIVE_BALANCE_INCREMENT {
		unslashedParticipatingTotalBalance = spec.EFFECTIVE_BALANCE_INCREMENT
	}
	unslashedParticipatingIncrements := unslashedParticipatingTotalBalance / spec.EFFECTIVE_BALANCE_INCREMENT

	activeIncrements := epc.TotalActiveStake / spec.EFFECTIVE_BALANCE_INCREMENT

	baseRewardPerIncrement := (spec.EFFECTIVE_BALANCE_INCREMENT * common.Gwei(spec.BASE_REWARD_FACTOR)) / epc.TotalActiveStakeSqRoot
	for _, vi := range attesterData.EligibleIndices {
		effBal := attesterData.Flats[vi].EffectiveBalance
		increments := effBal / spec.EFFECTIVE_BALANCE_INCREMENT
		baseReward := increments * baseRewardPerIncrement
		prevEpochParticipation := attesterData.PrevParticipation[vi]
		flagParticipation := prevEpochParticipation&flag != 0

		slashed := attesterData.Flats[vi].Slashed
		if !slashed && flagParticipation {
			if !isInactivityLeak {
				rewardNumerator := (baseReward * weight) * unslashedParticipatingIncrements
				rewardDenominator := activeIncrements * common.Gwei(WEIGHT_DENOMINATOR)
				out.Rewards[vi] += rewardNumerator / rewardDenominator
			}
		} else {
			out.Penalties[vi] += (baseReward * weight) / common.Gwei(WEIGHT_DENOMINATOR)
		}
	}
	return out, nil
}

func ComputeInactivityPenaltyDeltas(ctx context.Context, spec *common.Spec, epc *common.EpochsContext,
	attesterData *EpochAttesterData, inactivityScores *InactivityScoresView) (*common.Deltas, error) {
	out := common.NewDeltas(uint64(len(attesterData.Flats)))
	penaltyDenominator := common.Gwei(spec.INACTIVITY_SCORE_BIAS * spec.INACTIVITY_PENALTY_QUOTIENT_ALTAIR)
	for _, vi := range attesterData.EligibleIndices {
		if !(!attesterData.Flats[vi].Slashed && (attesterData.PrevParticipation[vi]&TIMELY_TARGET_FLAG != 0)) {
			score, err := inactivityScores.GetScore(vi)
			if err != nil {
				return nil, err
			}
			effBal := attesterData.Flats[vi].EffectiveBalance
			penaltyNumerator := effBal * common.Gwei(score)
			out.Penalties[vi] += penaltyNumerator / penaltyDenominator
		}
	}
	return out, nil
}

func ProcessEpochRewardsAndPenalties(ctx context.Context, spec *common.Spec, epc *common.EpochsContext,
	attesterData *EpochAttesterData, state *BeaconStateView) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	finalized, err := state.FinalizedCheckpoint()
	if err != nil {
		return err
	}
	finalityDelay := attesterData.PrevEpoch - finalized.Epoch
	isInactivityLeak := finalityDelay > spec.MIN_EPOCHS_TO_INACTIVITY_PENALTY

	currentEpoch := epc.CurrentEpoch.Epoch
	if currentEpoch == common.GENESIS_EPOCH {
		return nil
	}
	sourceDeltas, err := ComputeFlagDeltas(ctx, spec, epc, attesterData,
		TIMELY_SOURCE_FLAG, common.Gwei(TIMELY_SOURCE_WEIGHT), isInactivityLeak)
	if err != nil {
		return err
	}
	targetDeltas, err := ComputeFlagDeltas(ctx, spec, epc, attesterData,
		TIMELY_TARGET_FLAG, common.Gwei(TIMELY_TARGET_WEIGHT), isInactivityLeak)
	if err != nil {
		return err
	}
	headDeltas, err := ComputeFlagDeltas(ctx, spec, epc, attesterData,
		TIMELY_TARGET_FLAG, common.Gwei(TIMELY_TARGET_WEIGHT), isInactivityLeak)
	if err != nil {
		return err
	}
	inactivityScores, err := state.InactivityScores()
	if err != nil {
		return err
	}
	inactivityPenalties, err := ComputeInactivityPenaltyDeltas(ctx, spec, epc, attesterData, inactivityScores)
	if err != nil {
		return err
	}
	valCount := uint64(len(attesterData.Flats))
	sum := common.NewDeltas(valCount)
	sum.Add(sourceDeltas)
	sum.Add(targetDeltas)
	sum.Add(headDeltas)
	sum.Add(inactivityPenalties)
	balancesElements, err := common.ApplyDeltas(state, sum)
	if err != nil {
		return err
	}
	return state.setBalances(spec, balancesElements)
}
