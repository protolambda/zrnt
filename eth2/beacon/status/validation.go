package status

import (
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
)

// depends on ShufflingStatus
type ValidationStatus struct {
	ValidatorStatuses []AttesterStatus
}

func (status *ValidationStatus) Load(state *BeaconState, shufflingStatus *ShufflingStatus) {
	status.ValidatorStatuses = make([]AttesterStatus, len(state.Validators), len(state.Validators))

	previousBoundaryBlockRoot, _ := state.GetBlockRootAtSlot(state.PreviousEpoch().GetStartSlot())
	currentBoundaryBlockRoot, _ := state.GetBlockRootAtSlot(state.Epoch().GetStartSlot())

	participants := make([]ValidatorIndex, 0, MAX_VALIDATORS_PER_COMMITTEE)
	for _, att := range state.PreviousEpochAttestations {
		attBlockRoot, _ := state.GetBlockRootAtSlot(state.GetAttestationSlot(&att.Data))

		// attestation-target is already known to be previous-epoch, get it from the pre-computed shuffling directly.
		committee := shufflingStatus.Previous.Committees[att.Data.Crosslink.Shard]

		participants = participants[:0]                                     // reset old slice (re-used in for loop)
		participants = append(participants, committee...)                   // add committee indices
		participants = att.AggregationBits.FilterParticipants(participants) // only keep the participants
		for _, p := range participants {

			status := &status.ValidatorStatuses[p]

			// If the attestation is the earliest, i.e. has the biggest delay
			if status.InclusionDelay < att.InclusionDelay {
				status.InclusionDelay = att.InclusionDelay
				status.Proposer = att.ProposerIndex
			}

			if !state.Validators[p].Slashed {
				status.Flags |= UnslashedAttester
			}

			// remember the participant as one of the good validators
			status.Flags |= PrevEpochAttester

			// If the attestation is for the boundary:
			if att.Data.Target.Root == previousBoundaryBlockRoot {
				status.Flags |= PrevEpochBoundaryAttester
			}
			if att.Data.Target.Root == currentBoundaryBlockRoot {
				status.Flags |= CurrEpochBoundaryAttester
			}
			// If the attestation is for the head (att the time of attestation):
			if att.Data.BeaconBlockRoot == attBlockRoot {
				status.Flags |= MatchingHeadAttester
			}
		}
	}
	currentEpoch := state.Epoch()
	for i := 0; i < len(status.ValidatorStatuses); i++ {
		v := state.Validators[i]
		vStatus := &status.ValidatorStatuses[i]
		if v.IsActive(currentEpoch) || (v.Slashed && currentEpoch < v.WithdrawableEpoch) {
			vStatus.Flags |= EligibleAttester
		}
	}
	return
}
