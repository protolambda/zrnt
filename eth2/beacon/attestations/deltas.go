package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/math"
)

type AttestationDeltasFeature struct {
	Meta interface {
		meta.Versioning
		meta.RegistrySize
		meta.Staking
		meta.EffectiveBalances
		meta.AttesterStatuses
		meta.Finality
	}
}

func (f *AttestationDeltasFeature) AttestationDeltas() (*Deltas, error) {
	cres, err := f.Meta.ValidatorCount()
	if err != nil {
		return nil, err
	}
	validatorCount := ValidatorIndex(cres)
	deltas := NewDeltas(uint64(validatorCount))

	previousEpoch, err := f.Meta.PreviousEpoch()
	if err != nil {
		return nil, err
	}

	totalBalance, err := f.Meta.GetTotalStake()
	if err != nil {
		return nil, err
	}

	attesterStatuses, err := f.Meta.GetAttesterStatuses()
	if err != nil {
		return nil, err
	}
	prevEpochSourceStake, err := f.Meta.GetAttestersStake(attesterStatuses, PrevSourceAttester|UnslashedAttester)
	if err != nil {
		return nil, err
	}
	prevEpochTargetStake, err := f.Meta.GetAttestersStake(attesterStatuses, PrevTargetAttester|UnslashedAttester)
	if err != nil {
		return nil, err
	}
	prevEpochHeadStake, err := f.Meta.GetAttestersStake(attesterStatuses, PrevHeadAttester|UnslashedAttester)
	if err != nil {
		return nil, err
	}

	balanceSqRoot := Gwei(math.IntegerSquareroot(uint64(totalBalance)))
	finalized, err := f.Meta.Finalized()
	if err != nil {
		return nil, err
	}
	finalityDelay := previousEpoch - finalized.Epoch

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := attesterStatuses[i]
		if status.Flags&EligibleAttester != 0 {

			effBalance, err := f.Meta.EffectiveBalance(i)
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
