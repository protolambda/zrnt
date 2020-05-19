package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/math"
)

type AttestationRewardsAndPenaltiesFeature struct {
	Meta interface {
		meta.Versioning
		meta.RegistrySize
		meta.Staking
		meta.EffectiveBalances
		meta.AttesterStatuses
		meta.Finality
	}
}

func (f *AttestationRewardsAndPenaltiesFeature) AttestationRewardsAndPenalties() *RewardsAndPenalties {
	validatorCount := ValidatorIndex(f.Meta.ValidatorCount())

	res := NewRewardsAndPenalties(uint64(validatorCount))

	previousEpoch := f.Meta.PreviousEpoch()

	totalBalance := f.Meta.GetTotalActiveStake(f.Meta.CurrentEpoch())

	attesterStatuses := f.Meta.GetAttesterStatuses()
	prevEpochSourceStake := f.Meta.GetAttestersStake(attesterStatuses, PrevSourceAttester|UnslashedAttester)
	prevEpochTargetStake := f.Meta.GetAttestersStake(attesterStatuses, PrevTargetAttester|UnslashedAttester)
	prevEpochHeadStake := f.Meta.GetAttestersStake(attesterStatuses, PrevHeadAttester|UnslashedAttester)

	balanceSqRoot := Gwei(math.IntegerSquareroot(uint64(totalBalance)))
	finalityDelay := previousEpoch - f.Meta.Finalized().Epoch

	// All summed effective balances are normalized to effective-balance increments, to avoid overflows.
	totalBalance /= EFFECTIVE_BALANCE_INCREMENT
	prevEpochSourceStake /= EFFECTIVE_BALANCE_INCREMENT
	prevEpochTargetStake /= EFFECTIVE_BALANCE_INCREMENT
	prevEpochHeadStake /= EFFECTIVE_BALANCE_INCREMENT

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := attesterStatuses[i]

		effBalance := f.Meta.EffectiveBalance(i)
		baseReward := effBalance * BASE_REWARD_FACTOR /
			balanceSqRoot / BASE_REWARDS_PER_EPOCH

		// Inclusion delay
		if status.Flags.HasMarkers(PrevSourceAttester | UnslashedAttester) {
			// Inclusion speed bonus
			proposerReward := baseReward / PROPOSER_REWARD_QUOTIENT
			res.InclusionDelay.Rewards[status.AttestedProposer] += proposerReward
			maxAttesterReward := baseReward - proposerReward
			res.InclusionDelay.Rewards[i] += maxAttesterReward / Gwei(status.InclusionDelay)
		}

		if status.Flags&EligibleAttester != 0 {
			// Expected FFG source
			if status.Flags.HasMarkers(PrevSourceAttester | UnslashedAttester) {
				// Justification-participation reward
				res.Source.Rewards[i] += baseReward * prevEpochSourceStake / totalBalance
			} else {
				//Justification-non-participation R-penalty
				res.Source.Penalties[i] += baseReward
			}

			// Expected FFG target
			if status.Flags.HasMarkers(PrevTargetAttester | UnslashedAttester) {
				// Boundary-attestation reward
				res.Target.Rewards[i] += baseReward * prevEpochTargetStake / totalBalance
			} else {
				//Boundary-attestation-non-participation R-penalty
				res.Target.Penalties[i] += baseReward
			}

			// Expected head
			if status.Flags.HasMarkers(PrevHeadAttester | UnslashedAttester) {
				// Canonical-participation reward
				res.Head.Rewards[i] += baseReward * prevEpochHeadStake / totalBalance
			} else {
				// Non-canonical-participation R-penalty
				res.Head.Penalties[i] += baseReward
			}

			// Take away max rewards if we're not finalizing
			if finalityDelay > MIN_EPOCHS_TO_INACTIVITY_PENALTY {
				res.Inactivity.Penalties[i] += baseReward * BASE_REWARDS_PER_EPOCH
				if !status.Flags.HasMarkers(PrevTargetAttester|UnslashedAttester) {
					res.Inactivity.Penalties[i] += effBalance * Gwei(finalityDelay) / INACTIVITY_PENALTY_QUOTIENT
				}
			}
		}
	}

	return res
}
