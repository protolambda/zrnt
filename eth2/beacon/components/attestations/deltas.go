package attestations

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/meta"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
)

type AttestationDeltasReq interface {
	VersioningMeta
	RegistrySizeMeta
	StakingMeta
	AttesterStatusMeta
	FinalityMeta
}

func AttestationDeltas(meta AttestationDeltasReq) *Deltas {
	validatorCount := ValidatorIndex(meta.ValidatorCount())
	deltas := NewDeltas(uint64(validatorCount))

	previousEpoch := meta.PreviousEpoch()

	var totalBalance, totalAttestingBalance, epochBoundaryBalance, matchingHeadBalance Gwei
	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := meta.GetAttesterStatus(i)
		b := meta.EffectiveBalance(i)
		totalBalance += b
		if status.Flags.HasMarkers(PrevEpochAttester | UnslashedAttester) {
			totalAttestingBalance += b
		}
		if status.Flags.HasMarkers(PrevEpochBoundaryAttester | UnslashedAttester) {
			epochBoundaryBalance += b
		}
		if status.Flags.HasMarkers(MatchingHeadAttester | UnslashedAttester) {
			matchingHeadBalance += b
		}
	}
	previousTotalBalance := meta.GetTotalActiveEffectiveBalance(meta.PreviousEpoch())

	balanceSqRoot := Gwei(math.IntegerSquareroot(uint64(previousTotalBalance)))
	finalityDelay := previousEpoch - meta.Finalized().Epoch

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := meta.GetAttesterStatus(i)
		if status.Flags&EligibleAttester != 0 {

			effBalance := meta.EffectiveBalance(i)
			baseReward := effBalance * BASE_REWARD_FACTOR /
				balanceSqRoot / BASE_REWARDS_PER_EPOCH

			// Expected FFG source
			if status.Flags.HasMarkers(PrevEpochAttester | UnslashedAttester) {
				// Justification-participation reward
				deltas.Rewards[i] += baseReward * totalAttestingBalance / totalBalance

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
			if status.Flags.HasMarkers(PrevEpochBoundaryAttester | UnslashedAttester) {
				// Boundary-attestation reward
				deltas.Rewards[i] += baseReward * epochBoundaryBalance / totalBalance
			} else {
				//Boundary-attestation-non-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Expected head
			if status.Flags.HasMarkers(MatchingHeadAttester | UnslashedAttester) {
				// Canonical-participation reward
				deltas.Rewards[i] += baseReward * matchingHeadBalance / totalBalance
			} else {
				// Non-canonical-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Take away max rewards if we're not finalizing
			if finalityDelay > MIN_EPOCHS_TO_INACTIVITY_PENALTY {
				deltas.Penalties[i] += baseReward * BASE_REWARDS_PER_EPOCH
				if !status.Flags.HasMarkers(MatchingHeadAttester | UnslashedAttester) {
					deltas.Penalties[i] += effBalance * Gwei(finalityDelay) / INACTIVITY_PENALTY_QUOTIENT
				}
			}
		}
	}

	return deltas
}
