package deltas_computation

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/math"
)

type ValidatorStatusFlag uint64

func (flags ValidatorStatusFlag) hasMarkers(markers ValidatorStatusFlag) bool {
	return flags & markers == markers
}

const (
	prevEpochAttester ValidatorStatusFlag = 1 << iota
	matchingHeadAttester
	epochBoundaryAttester
	unslashed
	eligibleAttester
)

const noInclusionMarker = ^Slot(0)

type ValidatorStatus struct {
	InclusionSlot Slot
	DataSlot Slot
	Flags ValidatorStatusFlag
}

// Retrieves the inclusion slot of the earliest attestation in the previous epoch.
// Ok == true if there is such attestation, false otherwise.
func (status *ValidatorStatus) inclusionSlot() (slot Slot, ok bool) {
	if status.InclusionSlot == noInclusionMarker {
		return 0, false
	} else {
		return status.InclusionSlot, true
	}
}

// Note: ONLY safe to call when vIndex is known to be an active validator in the previous epoch
func (status *ValidatorStatus) inclusionDistance() Slot {
	return status.InclusionSlot - status.DataSlot
}

func DeltasJustificationAndFinalizationDeltas(state *BeaconState,) *Deltas {
	validatorCount := ValidatorIndex(len(state.ValidatorRegistry))
	deltas := NewDeltas(uint64(validatorCount))

	currentEpoch := state.Epoch()

	data := make([]ValidatorStatus, validatorCount, validatorCount)
	for i := 0; i < len(data); i++ {
		data[i].InclusionSlot = noInclusionMarker
	}

	{
		previousBoundaryBlockRoot, _ := state.GetBlockRoot(state.PreviousEpoch().GetStartSlot())

		for _, att := range state.PreviousEpochAttestations {
			attBlockRoot, _ := state.GetBlockRoot(att.Data.Slot)
			participants, _ := state.GetAttestationParticipants(&att.Data, &att.AggregationBitfield)
			for _, p := range participants {

				status := &data[p]

				// If the attestation is the earliest:
				if status.InclusionSlot > att.InclusionSlot {
					status.InclusionSlot = att.InclusionSlot
					status.DataSlot = att.Data.Slot
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
	}

	var totalBalance, totalAttestingBalance, epochBoundaryBalance, matchingHeadBalance Gwei
	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := &data[i]
		b := state.GetEffectiveBalance(i)
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
		v := state.ValidatorRegistry[i]
		if v.IsActive(currentEpoch) || (v.Slashed && currentEpoch < v.WithdrawableEpoch) {
			status.Flags |= eligibleAttester
		}
	}
	previousTotalBalance := state.GetTotalBalanceOf(
		state.ValidatorRegistry.GetActiveValidatorIndices(state.PreviousEpoch()))

	adjustedQuotient := math.IntegerSquareroot(uint64(previousTotalBalance)) / BASE_REWARD_QUOTIENT
	epochsSinceFinality := currentEpoch + 1 - state.FinalizedEpoch

	for i := ValidatorIndex(0); i < validatorCount; i++ {
		status := &data[i]
		if status.Flags & eligibleAttester != 0 {

			effectiveBalance := state.GetEffectiveBalance(i)
			baseReward := Gwei(0)
			if adjustedQuotient != 0 {
				baseReward = effectiveBalance / Gwei(adjustedQuotient) / 5
			}
			inactivityPenalty := baseReward
			if epochsSinceFinality > 4 {
				inactivityPenalty += effectiveBalance * Gwei(epochsSinceFinality) / INACTIVITY_PENALTY_QUOTIENT / 2
			}

			// Expected FFG source
			if status.Flags.hasMarkers(prevEpochAttester | unslashed) {
				// Justification-participation reward
				deltas.Rewards[i] += baseReward * totalAttestingBalance / totalBalance

				// Inclusion speed bonus
				inclusionDelay := status.inclusionDistance()
				deltas.Rewards[i] += baseReward * Gwei(MIN_ATTESTATION_INCLUSION_DELAY) / Gwei(inclusionDelay)
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
				deltas.Penalties[i] += inactivityPenalty
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
			if epochsSinceFinality > 4 {
				deltas.Penalties[i] += baseReward * 4
			}
		}
	}

	return deltas
}
