package deltas_computation

import (
	"github.com/protolambda/zrnt/eth2/beacon"
)

func DeltasJustification(state *beacon.BeaconState, v beacon.Valuator) *beacon.Deltas {
	epochsSinceFinality := state.Epoch() + 1 - state.FinalizedEpoch
	if epochsSinceFinality <= 4 {
		return ComputeNormalJustificationAndFinalization(state, v)
	} else {
		return ComputeInactivityLeakDeltas(state, v)
	}
}

type AttestersJustificationData struct {
	previousEpochEarliestAttestations map[beacon.ValidatorIndex]*beacon.PendingAttestation
	prevEpochAttesters []beacon.ValidatorIndex
	epochBoundaryAttesterIndices []beacon.ValidatorIndex
	matchingHeadAttesterIndices []beacon.ValidatorIndex
}

func (attJustData *AttestersJustificationData) inclusionSlot(vIndex beacon.ValidatorIndex) beacon.Slot {
	return attJustData.previousEpochEarliestAttestations[vIndex].InclusionSlot
}
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
			if att.Data.EpochBoundaryRoot == previousBoundaryBlockRoot {
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

// >> case 1: 4 or less epochs since finality
func ComputeNormalJustificationAndFinalization(state *beacon.BeaconState, v beacon.Valuator) *beacon.Deltas {
	deltas := beacon.NewDeltas(uint64(len(state.ValidatorRegistry)))

	data := GetAttestersJustificationData(state)
	totalAttestingBalance := state.ValidatorBalances.GetTotalBalance(data.prevEpochAttesters)
	totalBalance := state.ValidatorBalances.GetBalanceSum()
	epochBoundaryBalance := state.ValidatorBalances.GetTotalBalance(data.epochBoundaryAttesterIndices)
	matchingHeadBalance := state.ValidatorBalances.GetTotalBalance(data.matchingHeadAttesterIndices)

	prevActiveValidators := state.ValidatorRegistry.GetActiveValidatorIndices(state.PreviousEpoch())

	// Expected FFG source
	in, out := FindInAndOutValidators(prevActiveValidators, data.prevEpochAttesters)
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
	in, out = FindInAndOutValidators(prevActiveValidators, data.epochBoundaryAttesterIndices)
	for _, vIndex := range in {
		// Boundary-attestation reward
		deltas.Rewards[vIndex] += v.GetBaseReward(vIndex) * epochBoundaryBalance / totalBalance
	}
	for _, vIndex := range out {
		//Boundary-attestation-non-participation R-penalty
		deltas.Penalties[vIndex] += v.GetBaseReward(vIndex)
	}

	// Expected head
	in, out = FindInAndOutValidators(prevActiveValidators, data.matchingHeadAttesterIndices)
	for _, vIndex := range in {
		// Canonical-participation reward
		deltas.Rewards[vIndex] += v.GetBaseReward(vIndex) * matchingHeadBalance / totalBalance
	}
	for _, vIndex := range out {
		//Non-canonical-participation R-penalty
		deltas.Penalties[vIndex] += v.GetBaseReward(vIndex)
	}

	// Proposer bonus
	for _, vIndex := range prevActiveValidators {
		proposerIndex := state.GetBeaconProposerIndex(data.inclusionSlot(vIndex), false)
		deltas.Rewards[proposerIndex] += v.GetBaseReward(vIndex) / beacon.ATTESTATION_INCLUSION_REWARD_QUOTIENT
	}

	return deltas
}

// >> case 2: more than 4 epochs since finality. I.e. when blocks are not finalizing normally...
func ComputeInactivityLeakDeltas(state *beacon.BeaconState, v beacon.Valuator) *beacon.Deltas {
	deltas := beacon.NewDeltas(uint64(len(state.ValidatorRegistry)))

	data := GetAttestersJustificationData(state)
	prevActiveValidators := state.ValidatorRegistry.GetActiveValidatorIndices(state.PreviousEpoch())
	in, out := FindInAndOutValidators(prevActiveValidators, data.prevEpochAttesters)
	for _, vIndex := range in {
		// Attestation delay measure
		// If a validator did attest, apply a small penalty for getting attestations included late
		deltas.Rewards[vIndex] += v.GetBaseReward(vIndex) *
			beacon.Gwei(beacon.MIN_ATTESTATION_INCLUSION_DELAY) / beacon.Gwei(data.inclusionDistance(vIndex))
		deltas.Penalties[vIndex] += v.GetBaseReward(vIndex)
	}
	for _, vIndex := range out {
		// Justification-inactivity penalty
		deltas.Penalties[vIndex] += v.GetInactivityPenalty(vIndex)
	}

	_, out = FindInAndOutValidators(prevActiveValidators, data.epochBoundaryAttesterIndices)
	for _, vIndex := range out {
		// Boundary-attestation-Inactivity penalty
		deltas.Penalties[vIndex] += v.GetInactivityPenalty(vIndex)
	}

	_, out = FindInAndOutValidators(prevActiveValidators, data.matchingHeadAttesterIndices)
	for _, vIndex := range out {
		// Non-canonical-participation R-penalty
		deltas.Penalties[vIndex] += v.GetBaseReward(vIndex)
	}

	// Find the validators that are inactive and misbehaving
	eligible := make([]beacon.ValidatorIndex, len(state.ValidatorRegistry))
	for i := 0; i < len(state.ValidatorRegistry); i++ {
		v := &state.ValidatorRegistry[i]
		// Eligible if slashed and not withdrawn
		if v.Slashed && state.Epoch() < v.WithdrawableEpoch {
			eligible[i] = beacon.ValidatorIndex(i)
		} else {
			eligible[i] = beacon.ValidatorIndexMarker
		}
	}
	// exclude all indices that are not eligible because of their activity
	for _, vIndex := range prevActiveValidators {
		eligible[vIndex] = beacon.ValidatorIndexMarker
	}

	// Slash all eligible inactive misbehaving validators.
	for _, vIndex := range eligible {
		if vIndex != beacon.ValidatorIndexMarker {
			// Penalization measure: double inactivity penalty + R-penalty
			deltas.Penalties[vIndex] += 2 * v.GetInactivityPenalty(vIndex) + v.GetBaseReward(vIndex)
		}
	}

	return deltas
}
