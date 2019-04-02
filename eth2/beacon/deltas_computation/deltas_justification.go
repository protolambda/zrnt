package deltas_computation

import (
	"github.com/protolambda/zrnt/eth2/beacon"
)

type AttestersJustificationData struct {
	previousEpochEarliestAttestations map[beacon.ValidatorIndex]*beacon.PendingAttestation
	prevEpochAttesters                []beacon.ValidatorIndex
	epochBoundaryAttesterIndices      []beacon.ValidatorIndex
	matchingHeadAttesterIndices       []beacon.ValidatorIndex
}

// Retrieves the inclusion slot of the earliest attestation in the previous epoch by the given vIndex.
// Ok == true if there is such attestation, false otherwise.
func (attJustData *AttestersJustificationData) inclusionSlot(vIndex beacon.ValidatorIndex) (slot beacon.Slot, ok bool) {
	if att, ok := attJustData.previousEpochEarliestAttestations[vIndex]; ok {
		return att.InclusionSlot, true
	} else {
		return 0, false
	}
}

// Note: ONLY safe to call when vIndex is known to be an active validator in the previous epoch
func (attJustData *AttestersJustificationData) inclusionDistance(vIndex beacon.ValidatorIndex) beacon.Slot {
	a := attJustData.previousEpochEarliestAttestations[vIndex]
	return a.InclusionSlot - a.Data.Slot
}

func GetAttestersJustificationData(state *beacon.BeaconState) *AttestersJustificationData {
	data := new(AttestersJustificationData)

	prevEpochAttestersSet := make(map[beacon.ValidatorIndex]struct{}, 0)

	// attestation-source-index for a given epoch, by validator index.
	// The earliest attestation (by inclusion_slot) is referenced in this map.
	data.previousEpochEarliestAttestations = make(map[beacon.ValidatorIndex]*beacon.PendingAttestation)

	previousBoundaryBlockRoot, _ := state.GetBlockRoot(state.PreviousEpoch().GetStartSlot())

	data.epochBoundaryAttesterIndices = make([]beacon.ValidatorIndex, 0)
	data.matchingHeadAttesterIndices = make([]beacon.ValidatorIndex, 0)
	for i := 0; i < len(state.PreviousEpochAttestations); i++ {
		att := &state.PreviousEpochAttestations[i]
		attBlockRoot, _ := state.GetBlockRoot(att.Data.Slot)
		participants, _ := state.GetAttestationParticipants(&att.Data, &att.AggregationBitfield)
		for _, p := range participants {

			// remember the participant as one of the good validators
			prevEpochAttestersSet[p] = struct{}{}

			// If the attestation is the earliest:
			if existingAtt, ok := data.previousEpochEarliestAttestations[p];
				!ok || existingAtt.InclusionSlot < att.InclusionSlot {
				data.previousEpochEarliestAttestations[p] = att
			}

			// If the attestation is for the boundary:
			if att.Data.TargetRoot == previousBoundaryBlockRoot {
				data.epochBoundaryAttesterIndices = append(data.epochBoundaryAttesterIndices, p)
			}
			// If the attestation is for the head (att the time of attestation):
			if att.Data.BeaconBlockRoot == attBlockRoot {
				data.matchingHeadAttesterIndices = append(data.matchingHeadAttesterIndices, p)
			}
		}

	}
	data.prevEpochAttesters = make([]beacon.ValidatorIndex, len(prevEpochAttestersSet), len(prevEpochAttestersSet))
	i := 0
	for vIndex := range prevEpochAttestersSet {
		data.prevEpochAttesters[i] = vIndex
		i++
	}
	return data
}

func DeltasJustificationAndFinalizationDeltas(state *beacon.BeaconState, v beacon.Valuator) *beacon.Deltas {
	deltas := beacon.NewDeltas(uint64(len(state.ValidatorRegistry)))

	currentEpoch := state.Epoch()

	data := GetAttestersJustificationData(state)
	totalAttestingBalance := state.GetTotalBalanceOf(data.prevEpochAttesters)
	totalBalance := state.GetTotalBalance()
	epochBoundaryBalance := state.GetTotalBalanceOf(data.epochBoundaryAttesterIndices)
	matchingHeadBalance := state.GetTotalBalanceOf(data.matchingHeadAttesterIndices)

	validatorCount := beacon.ValidatorIndex(len(state.ValidatorRegistry))
	eligibleValidators := make([]beacon.ValidatorIndex, 0, validatorCount)
	for i := beacon.ValidatorIndex(0); i < validatorCount; i++ {
		v := &state.ValidatorRegistry[i]
		if v.IsActive(currentEpoch) {
			eligibleValidators = append(eligibleValidators, i)
		} else if v.Slashed && currentEpoch < v.WithdrawableEpoch {
			eligibleValidators = append(eligibleValidators, i)
		}
	}

	// Expected FFG source
	in, out := beacon.FindInAndOutValidators(eligibleValidators, data.prevEpochAttesters)
	for _, vIndex := range in {
		// Justification-participation reward
		deltas.Rewards[vIndex] += v.GetBaseReward(vIndex) * totalAttestingBalance / totalBalance

		// Attestation-Inclusion-delay reward: quicker = more reward
		deltas.Rewards[vIndex] += v.GetBaseReward(vIndex) *
			beacon.Gwei(beacon.MIN_ATTESTATION_INCLUSION_DELAY) / beacon.Gwei(data.inclusionDistance(vIndex))
	}
	for _, vIndex := range out {
		//Justification-non-participation R-penalty
		deltas.Penalties[vIndex] += v.GetBaseReward(vIndex)
	}

	// Expected FFG target
	in, out = beacon.FindInAndOutValidators(eligibleValidators, data.epochBoundaryAttesterIndices)
	for _, vIndex := range in {
		// Boundary-attestation reward
		deltas.Rewards[vIndex] += v.GetBaseReward(vIndex) * epochBoundaryBalance / totalBalance
	}
	for _, vIndex := range out {
		//Boundary-attestation-non-participation R-penalty
		deltas.Penalties[vIndex] += v.GetBaseReward(vIndex)
	}

	// Expected head
	in, out = beacon.FindInAndOutValidators(eligibleValidators, data.matchingHeadAttesterIndices)
	for _, vIndex := range in {
		// Canonical-participation reward
		deltas.Rewards[vIndex] += v.GetBaseReward(vIndex) * matchingHeadBalance / totalBalance
	}
	for _, vIndex := range out {
		// Non-canonical-participation R-penalty
		deltas.Penalties[vIndex] += v.GetBaseReward(vIndex)
	}

	// Proposer bonus
	in, _ = beacon.FindInAndOutValidators(eligibleValidators, data.prevEpochAttesters)
	for _, vIndex := range in {
		inclSlot, ok := data.inclusionSlot(vIndex)
		if !ok {
			// active validator did not have an attestation included
			continue
		}
		proposerIndex := state.GetBeaconProposerIndex(inclSlot)
		deltas.Rewards[proposerIndex] += v.GetBaseReward(vIndex) / beacon.ATTESTATION_INCLUSION_REWARD_QUOTIENT
	}

	if v.IsNotFinalizing() {
		// Take away max rewards if we're not finalizing
		for _, vIndex := range eligibleValidators {
			deltas.Penalties[vIndex] += v.GetBaseReward(vIndex) * 4
		}
	}

	return deltas
}
