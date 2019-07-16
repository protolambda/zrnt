package components

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components/registry"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zssz/bitfields"
)

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

func (cb CommitteeBits) BitLen() uint64 {
	return bitfields.BitlistLen(cb)
}

func (cb CommitteeBits) GetBit(i uint64) bool {
	return bitfields.GetBit(cb, i)
}

func (cb *CommitteeBits) Limit() uint64 {
	return MAX_VALIDATORS_PER_COMMITTEE
}

// Sets the bits to true that are true in other. (in place)
func (cb CommitteeBits) Or(other CommitteeBits) {
	for i := 0; i < len(cb); i++ {
		cb[i] |= other[i]
	}
}

// In-place filters a list of committees indices to only keep the bitfield participants.
// The result is not sorted. Returns the re-sliced filtered participants list.
func (cb CommitteeBits) FilterParticipants(committee []ValidatorIndex) []ValidatorIndex {
	bitLen := cb.BitLen()
	out := committee[:0]
	if bitLen != uint64(len(committee)) {
		panic("committee mismatch, bitfield length does not match")
	}
	for i := uint64(0); i < bitLen; i++ {
		if cb.GetBit(i) {
			out = append(out, committee[i])
		}
	}
	return out
}

// In-place filters a list of committees indices to only keep the bitfield NON-participants.
// The result is not sorted. Returns the re-sliced filtered non-participants list.
func (cb CommitteeBits) FilterNonParticipants(committee []ValidatorIndex) []ValidatorIndex {
	bitLen := cb.BitLen()
	out := committee[:0]
	if bitLen != uint64(len(committee)) {
		panic("committee mismatch, bitfield length does not match")
	}
	for i := uint64(0); i < bitLen; i++ {
		if !cb.GetBit(i) {
			out = append(out, committee[i])
		}
	}
	return out
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

func (state *BeaconState) GetAttestationSlot(attData *AttestationData) Slot {
	epoch := attData.Target.Epoch
	committeeCount := Slot(state.Validators.GetCommitteeCount(epoch))
	offset := Slot((attData.Crosslink.Shard + SHARD_COUNT - state.GetStartShard(epoch)) % SHARD_COUNT)
	return epoch.GetStartSlot() + (offset / (committeeCount / SLOTS_PER_EPOCH))
}

type ValidatorStatusFlag uint64

func (flags ValidatorStatusFlag) HasMarkers(markers ValidatorStatusFlag) bool {
	return flags&markers == markers
}

const (
	PrevEpochAttester ValidatorStatusFlag = 1 << iota
	MatchingHeadAttester
	PrevEpochBoundaryAttester
	CurrEpochBoundaryAttester
	UnslashedAttester
	EligibleAttester
)

type ValidatorStatus struct {
	// no delay (i.e. 0) by default
	InclusionDelay Slot
	Proposer       ValidatorIndex
	Flags          ValidatorStatusFlag
}

type ValidationData interface {
	GetValidatorStatus(index ValidatorIndex) ValidatorStatus
}

func (state *BeaconState) AttestationDeltas() *Deltas {
	validatorCount := ValidatorIndex(len(state.Validators))
	deltas := NewDeltas(uint64(validatorCount))

	previousEpoch := state.PreviousEpoch()

	var totalBalance, totalAttestingBalance, epochBoundaryBalance, matchingHeadBalance Gwei
	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := state.PrecomputedData.GetValidatorStatus(i)
		v := state.Validators[i]
		b := v.EffectiveBalance
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
	previousTotalBalance := state.Validators.GetTotalActiveEffectiveBalance(state.PreviousEpoch())

	balanceSqRoot := Gwei(math.IntegerSquareroot(uint64(previousTotalBalance)))
	finalityDelay := previousEpoch - state.FinalizedCheckpoint.Epoch

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := state.PrecomputedData.GetValidatorStatus(i)
		if status.Flags&EligibleAttester != 0 {

			v := state.Validators[i]
			baseReward := v.EffectiveBalance * BASE_REWARD_FACTOR /
				balanceSqRoot / BASE_REWARDS_PER_EPOCH

			// Expected FFG source
			if status.Flags.HasMarkers(PrevEpochAttester | UnslashedAttester) {
				// Justification-participation reward
				deltas.Rewards[i] += baseReward * totalAttestingBalance / totalBalance

				// Inclusion speed bonus
				proposerReward := baseReward / PROPOSER_REWARD_QUOTIENT
				deltas.Rewards[status.Proposer] += proposerReward
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
					deltas.Penalties[i] += v.EffectiveBalance * Gwei(finalityDelay) / INACTIVITY_PENALTY_QUOTIENT
				}
			}
		}
	}

	return deltas
}
