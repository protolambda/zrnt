package components

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zssz"
)

var AttestationDataAndCustodyBitSSZ = zssz.GetSSZ((*AttestationDataAndCustodyBit)(nil))

type AttestationDataAndCustodyBit struct {
	// Attestation data
	Data AttestationData
	// Custody bit
	CustodyBit bool
}

type AttestationData struct {
	// Root of the signed beacon block
	BeaconBlockRoot Root

	// FFG vote
	SourceEpoch Epoch
	SourceRoot  Root
	TargetEpoch Epoch
	TargetRoot  Root

	// Crosslink vote
	Crosslink Crosslink
}

type PendingAttestation struct {
	// Attester aggregation bitfield
	AggregationBitfield bitfield.Bitfield
	// Attestation data
	Data AttestationData
	// Inclusion delay
	InclusionDelay Slot
	// Proposer index
	ProposerIndex ValidatorIndex
}

type AttestationsState struct {
	PreviousEpochAttestations []*PendingAttestation
	CurrentEpochAttestations  []*PendingAttestation
}

func (state *BeaconState) GetAttesters(attestations []*PendingAttestation, filter func(att *AttestationData) bool) ValidatorSet {
	out := make(ValidatorSet, 0)
	for _, att := range attestations {
		// If the attestation is for the boundary:
		if filter(&att.Data) {
			participants, _ := state.GetAttestingIndicesUnsorted(&att.Data, &att.AggregationBitfield)
			out = append(out, participants...)
		}
	}
	out.Dedup()
	return out
}

func (state *BeaconState) GetAttestationSlot(attData *AttestationData) Slot {
	epoch := attData.TargetEpoch
	committeeCount := Slot(state.Validators.GetEpochCommitteeCount(epoch))
	offset := Slot((attData.Crosslink.Shard + SHARD_COUNT - state.GetEpochStartShard(epoch)) % SHARD_COUNT)
	return epoch.GetStartSlot() + (offset / (committeeCount / SLOTS_PER_EPOCH))
}

func (state *BeaconState) AttestationDeltas() *Deltas {
	validatorCount := ValidatorIndex(len(state.Validators))
	deltas := NewDeltas(uint64(validatorCount))

	previousEpoch := state.PreviousEpoch()

	data := state.ValidationStatus()

	var totalBalance, totalAttestingBalance, epochBoundaryBalance, matchingHeadBalance Gwei
	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := &data[i]
		v := state.Validators[i]
		b := v.EffectiveBalance
		totalBalance += b
		if status.Flags.hasMarkers(PrevEpochAttester | UnslashedAttester) {
			totalAttestingBalance += b
		}
		if status.Flags.hasMarkers(EpochBoundaryAttester | UnslashedAttester) {
			epochBoundaryBalance += b
		}
		if status.Flags.hasMarkers(MatchingHeadAttester | UnslashedAttester) {
			matchingHeadBalance += b
		}
	}
	previousTotalBalance := state.Validators.GetTotalActiveEffectiveBalance(state.PreviousEpoch())

	balanceSqRoot := Gwei(math.IntegerSquareroot(uint64(previousTotalBalance)))
	finalityDelay := previousEpoch - state.FinalizedEpoch

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := &data[i]
		if status.Flags&EligibleAttester != 0 {

			v := state.Validators[i]
			baseReward := v.EffectiveBalance * BASE_REWARD_FACTOR /
				balanceSqRoot / BASE_REWARDS_PER_EPOCH

			// Expected FFG source
			if status.Flags.hasMarkers(PrevEpochAttester | UnslashedAttester) {
				// Justification-participation reward
				deltas.Rewards[i] += baseReward * totalAttestingBalance / totalBalance

				// Inclusion speed bonus
				deltas.Rewards[status.Proposer] += baseReward / PROPOSER_REWARD_QUOTIENT
				deltas.Rewards[i] += baseReward * Gwei(MIN_ATTESTATION_INCLUSION_DELAY) / Gwei(status.InclusionDelay)
			} else {
				//Justification-non-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Expected FFG target
			if status.Flags.hasMarkers(EpochBoundaryAttester | UnslashedAttester) {
				// Boundary-attestation reward
				deltas.Rewards[i] += baseReward * epochBoundaryBalance / totalBalance
			} else {
				//Boundary-attestation-non-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Expected head
			if status.Flags.hasMarkers(MatchingHeadAttester | UnslashedAttester) {
				// Canonical-participation reward
				deltas.Rewards[i] += baseReward * matchingHeadBalance / totalBalance
			} else {
				// Non-canonical-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Take away max rewards if we're not finalizing
			if finalityDelay > MIN_EPOCHS_TO_INACTIVITY_PENALTY {
				deltas.Penalties[i] += baseReward * BASE_REWARDS_PER_EPOCH
				if !status.Flags.hasMarkers(MatchingHeadAttester | UnslashedAttester) {
					deltas.Penalties[i] += v.EffectiveBalance * Gwei(finalityDelay) / INACTIVITY_PENALTY_QUOTIENT
				}
			}
		}
	}

	return deltas
}
