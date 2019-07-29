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

func (f *AttestationDeltasFeature) AttestationDeltas() *Deltas {
	validatorCount := ValidatorIndex(f.Meta.ValidatorCount())
	deltas := NewDeltas(uint64(validatorCount))

	previousEpoch := f.Meta.PreviousEpoch()

	totalBalance := f.Meta.GetTotalStake()
	prevEpochStake := f.Meta.GetTotalEpochStake(previousEpoch)

	balanceSqRoot := Gwei(math.IntegerSquareroot(uint64(totalBalance)))
	finalityDelay := previousEpoch - f.Meta.Finalized().Epoch

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := f.Meta.GetAttesterStatus(i)
		if status.Flags&EligibleAttester != 0 {

			effBalance := f.Meta.EffectiveBalance(i)
			baseReward := effBalance * BASE_REWARD_FACTOR /
				balanceSqRoot / BASE_REWARDS_PER_EPOCH

			// Expected FFG source
			if status.Flags.HasMarkers(PrevSourceAttester | UnslashedAttester) {
				// Justification-participation reward
				deltas.Rewards[i] += baseReward * prevEpochStake.SourceBalance / totalBalance

				// Inclusion speed bonus
				proposerReward := baseReward / PROPOSER_REWARD_QUOTIENT
				deltas.Rewards[status.AttestedProposer] += proposerReward
				maxAttesterReward := baseReward - proposerReward
				inclusionOffset := SLOTS_PER_EPOCH + MIN_ATTESTATION_INCLUSION_DELAY - status.InclusionDelay
				deltas.Rewards[i] += maxAttesterReward * Gwei(inclusionOffset) / Gwei(SLOTS_PER_EPOCH)
			} else {
				//Justification-non-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Expected FFG target
			if status.Flags.HasMarkers(PrevTargetAttester | UnslashedAttester) {
				// Boundary-attestation reward
				deltas.Rewards[i] += baseReward * prevEpochStake.TargetBalance / totalBalance
			} else {
				//Boundary-attestation-non-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Expected head
			if status.Flags.HasMarkers(PrevHeadAttester | UnslashedAttester) {
				// Canonical-participation reward
				deltas.Rewards[i] += baseReward * prevEpochStake.HeadBalance / totalBalance
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

	return deltas
}
