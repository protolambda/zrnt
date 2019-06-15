package deltas_computation

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zrnt/presets/generated"
)

type ValidatorStatusFlag uint64

func (flags ValidatorStatusFlag) hasMarkers(markers ValidatorStatusFlag) bool {
	return flags&markers == markers
}

const (
	prevEpochAttester ValidatorStatusFlag = 1 << iota
	matchingHeadAttester
	epochBoundaryAttester
	unslashed
	eligibleAttester
)

type ValidatorStatus struct {
	// no delay (i.e. 0) by default
	InclusionDelay Slot
	Proposer       ValidatorIndex
	Flags          ValidatorStatusFlag
}

type ValidatorStatusList []ValidatorStatus

func (vsl ValidatorStatusList) loadStatuses(state *BeaconState) {
	previousBoundaryBlockRoot, _ := state.GetBlockRootAtSlot(state.PreviousEpoch().GetStartSlot())

	for _, att := range state.PreviousEpochAttestations {
		attBlockRoot, _ := state.GetBlockRootAtSlot(state.GetAttestationSlot(&att.Data))
		participants, _ := state.GetAttestingIndicesUnsorted(&att.Data, &att.AggregationBitfield)
		for _, p := range participants {

			status := &vsl[p]

			// If the attestation is the earliest, i.e. has the biggest delay
			if status.InclusionDelay < att.InclusionDelay {
				status.InclusionDelay = att.InclusionDelay
				status.Proposer = att.ProposerIndex
			}

			if !state.ValidatorRegistry[p].Slashed {
				status.Flags |= unslashed
			}

			// remember the participant as one of the good validators
			status.Flags |= prevEpochAttester

			// If the attestation is for the boundary:
			if att.Data.TargetRoot == previousBoundaryBlockRoot {
				status.Flags |= epochBoundaryAttester
			}
			// If the attestation is for the head (att the time of attestation):
			if att.Data.BeaconBlockRoot == attBlockRoot {
				status.Flags |= matchingHeadAttester
			}
		}
	}
	currentEpoch := state.Epoch()
	for i := 0; i < len(vsl); i++ {
		v := state.ValidatorRegistry[i]
		status := &vsl[i]
		if v.IsActive(currentEpoch) || (v.Slashed && currentEpoch < v.WithdrawableEpoch) {
			status.Flags |= eligibleAttester
		}
	}
}

func AttestationDeltas(state *BeaconState) *Deltas {
	validatorCount := ValidatorIndex(len(state.ValidatorRegistry))
	deltas := NewDeltas(uint64(validatorCount))

	previousEpoch := state.PreviousEpoch()

	data := make(ValidatorStatusList, validatorCount, validatorCount)
	data.loadStatuses(state)

	var totalBalance, totalAttestingBalance, epochBoundaryBalance, matchingHeadBalance Gwei
	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := &data[i]
		v := state.ValidatorRegistry[i]
		b := v.EffectiveBalance
		totalBalance += b
		if status.Flags.hasMarkers(prevEpochAttester | unslashed) {
			totalAttestingBalance += b
		}
		if status.Flags.hasMarkers(epochBoundaryAttester | unslashed) {
			epochBoundaryBalance += b
		}
		if status.Flags.hasMarkers(matchingHeadAttester | unslashed) {
			matchingHeadBalance += b
		}
	}
	previousTotalBalance := state.GetTotalBalanceOf(
		state.ValidatorRegistry.GetActiveValidatorIndices(state.PreviousEpoch()))

	balanceSqRoot := Gwei(math.IntegerSquareroot(uint64(previousTotalBalance)))
	finalityDelay := previousEpoch - state.FinalizedEpoch

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := &data[i]
		if status.Flags&eligibleAttester != 0 {

			v := state.ValidatorRegistry[i]
			baseReward := v.EffectiveBalance * BASE_REWARD_FACTOR /
				balanceSqRoot / BASE_REWARDS_PER_EPOCH


			// Expected FFG source
			if status.Flags.hasMarkers(prevEpochAttester | unslashed) {
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
			if status.Flags.hasMarkers(epochBoundaryAttester | unslashed) {
				// Boundary-attestation reward
				deltas.Rewards[i] += baseReward * epochBoundaryBalance / totalBalance
			} else {
				//Boundary-attestation-non-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Expected head
			if status.Flags.hasMarkers(matchingHeadAttester | unslashed) {
				// Canonical-participation reward
				deltas.Rewards[i] += baseReward * matchingHeadBalance / totalBalance
			} else {
				// Non-canonical-participation R-penalty
				deltas.Penalties[i] += baseReward
			}

			// Take away max rewards if we're not finalizing
			if finalityDelay > constant_presets.MIN_EPOCHS_TO_INACTIVITY_PENALTY {
				deltas.Penalties[i] += baseReward * BASE_REWARDS_PER_EPOCH
				if !status.Flags.hasMarkers(matchingHeadAttester | unslashed) {
					deltas.Penalties[i] += v.EffectiveBalance * Gwei(finalityDelay) / INACTIVITY_PENALTY_QUOTIENT
				}
			}
		}
	}

	return deltas
}
