package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/math"
)

type AttestationDeltasInput interface {
	meta.Versioning
	meta.RegistrySize
	meta.Staking
	meta.DetailedStaking
	meta.EffectiveBalances
	meta.Finality
}

func AttestationDeltas(input AttestationDeltasInput) (*Deltas, error) {
	cres, err := input.ValidatorCount()
	if err != nil {
		return nil, err
	}
	validatorCount := ValidatorIndex(cres)
	deltas := NewDeltas(uint64(validatorCount))

	previousEpoch, err := input.PreviousEpoch()
	if err != nil {
		return nil, err
	}

	attesterStatuses, err := input.GetAttesterStatuses()
	if err != nil {
		return nil, err
	}

	totalBalance, err := input.GetTotalStake()
	if err != nil {
		return nil, err
	}

	prevEpochStake, err := input.PrevEpochStakeSummary()
	if err != nil {
		return nil, err
	}
	prevEpochSourceStake := prevEpochStake.SourceStake
	prevEpochTargetStake := prevEpochStake.TargetStake
	prevEpochHeadStake := prevEpochStake.HeadStake

	balanceSqRoot := Gwei(math.IntegerSquareroot(uint64(totalBalance)))
	finalized, err := input.Finalized()
	if err != nil {
		return nil, err
	}
	finalityDelay := previousEpoch - finalized.Epoch

	// All summed effective balances are normalized to effective-balance increments, to avoid overflows.
	totalBalance /= EFFECTIVE_BALANCE_INCREMENT
	prevEpochSourceStake /= EFFECTIVE_BALANCE_INCREMENT
	prevEpochTargetStake /= EFFECTIVE_BALANCE_INCREMENT
	prevEpochHeadStake /= EFFECTIVE_BALANCE_INCREMENT

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := attesterStatuses[i]
		if status.Flags&EligibleAttester != 0 {

			effBalance, err := input.EffectiveBalance(i)
			if err != nil {
				return nil, err
			}
			baseReward := effBalance * BASE_REWARD_FACTOR /
				balanceSqRoot / BASE_REWARDS_PER_EPOCH

			// Expected FFG source
			if status.Flags.HasMarkers(PrevSourceAttester | UnslashedAttester) {
				// Justification-participation reward
				deltas.Rewards[i] += baseReward * prevEpochSourceStake / totalBalance

				// Inclusion speed bonus
				proposerReward := baseReward / PROPOSER_REWARD_QUOTIENT
				deltas.Rewards[status.AttestedProposer] += proposerReward
				maxAttesterReward := baseReward - proposerReward
				deltas.Rewards[i] += maxAttesterReward / Gwei(status.InclusionDelay)
			} else {
				//Justification-non-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Expected FFG target
			if status.Flags.HasMarkers(PrevTargetAttester | UnslashedAttester) {
				// Boundary-attestation reward
				deltas.Rewards[i] += baseReward * prevEpochTargetStake / totalBalance
			} else {
				//Boundary-attestation-non-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Expected head
			if status.Flags.HasMarkers(PrevHeadAttester | UnslashedAttester) {
				// Canonical-participation reward
				deltas.Rewards[i] += baseReward * prevEpochHeadStake / totalBalance
			} else {
				// Non-canonical-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Take away max rewards if we're not finalizing
			if finalityDelay > MIN_EPOCHS_TO_INACTIVITY_PENALTY {
				deltas.Penalties[i] += baseReward * BASE_REWARDS_PER_EPOCH
				if !status.Flags.HasMarkers(PrevHeadAttester | UnslashedAttester) {
					deltas.Penalties[i] += effBalance * Gwei(finalityDelay) / INACTIVITY_PENALTY_QUOTIENT
				}
			}
		}
	}

	return deltas, nil
}
