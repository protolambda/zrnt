package components

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zssz"
	"github.com/protolambda/zssz/bitfields"
)

var AttestationDataAndCustodyBitSSZ = zssz.GetSSZ((*AttestationDataAndCustodyBit)(nil))

type AttestationDataAndCustodyBit struct {
	Data       AttestationData
	CustodyBit bool // Challengeable bit (SSZ-bool, 1 byte) for the custody of crosslink data
}

type AttestationData struct {
	// LMD GHOST vote
	BeaconBlockRoot Root

	// FFG vote
	Source Checkpoint
	Target Checkpoint

	// Crosslink vote
	Crosslink Crosslink
}

type CommitteeBits []byte

func (cb CommitteeBits) BitLen() uint32 {
	return bitfields.BitlistLen(cb)
}

func (cb *CommitteeBits) Limit() uint32 {
	return MAX_VALIDATORS_PER_COMMITTEE
}

type PendingAttestation struct {
	AggregationBits CommitteeBits
	Data            AttestationData
	InclusionDelay  Slot
	ProposerIndex   ValidatorIndex
}

type AttestationsState struct {
	PreviousEpochAttestations []*PendingAttestation
	CurrentEpochAttestations  []*PendingAttestation
}

// Rotate current/previous epoch attestations
func (state *AttestationsState) RotateEpochAttestations() {
	state.PreviousEpochAttestations = state.CurrentEpochAttestations
	state.CurrentEpochAttestations = nil
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
