package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
)

type AttesterStatuses []AttesterStatus

func (ats AttesterStatuses) GetAttesterStatus(index ValidatorIndex) AttesterStatus {
	return ats[index]
}

type AttesterStatusReq interface {
	VersioningMeta
	RegistrySizeMeta
	CrosslinkTimingMeta
	CommitteeCountMeta
	CrosslinkCommitteeMeta
	EffectiveBalanceMeta
	HistoryMeta
	SlashedMeta
	ActiveIndicesMeta
}

func (state *AttestationsState) LoadAttesterStatuses(meta AttesterStatusReq) (out AttesterStatuses) {
	count := meta.ValidatorCount()
	out = make(AttesterStatuses, count, count)

	currentEpoch := meta.Epoch()
	prevEpoch := meta.PreviousEpoch()
	previousBoundaryBlockRoot := meta.GetBlockRootAtSlot(prevEpoch.GetStartSlot())
	//currentBoundaryBlockRoot := meta.GetBlockRootAtSlot(currentEpoch.GetStartSlot())

	participants := make([]ValidatorIndex, 0, MAX_VALIDATORS_PER_COMMITTEE)
	for _, att := range state.PreviousEpochAttestations {
		attBlockRoot := meta.GetBlockRootAtSlot(state.GetAttestationSlot(meta, &att.Data))

		// attestation-target is already known to be previous-epoch, get it from the pre-computed shuffling directly.
		committee := meta.GetCrosslinkCommittee(prevEpoch, att.Data.Crosslink.Shard)

		participants = participants[:0]                                     // reset old slice (re-used in for loop)
		participants = append(participants, committee...)                   // add committee indices
		participants = att.AggregationBits.FilterParticipants(participants) // only keep the participants
		for _, p := range participants {

			status := &out[p]

			// If the attestation is the earliest, i.e. has the biggest delay
			if status.InclusionDelay < att.InclusionDelay {
				status.InclusionDelay = att.InclusionDelay
				status.AttestedProposer = att.ProposerIndex
			}

			if !meta.IsSlashed(p) {
				status.Flags |= UnslashedAttester
			}

			// remember the participant as one of the good validators
			status.Flags |= PrevEpochAttester

			// If the attestation is for the boundary:
			if att.Data.Target.Root == previousBoundaryBlockRoot {
				status.Flags |= PrevEpochBoundaryAttester
			}
			// If the attestation is for the head (att the time of attestation):
			if att.Data.BeaconBlockRoot == attBlockRoot {
				status.Flags |= MatchingHeadAttester
			}
		}
	}
	// TODO
	//if att.Data.Target.Root == currentBoundaryBlockRoot {
	//	status.Flags |= CurrEpochBoundaryAttester
	//}
	for _, index := range meta.GetActiveValidatorIndices(currentEpoch) {
		out[index].Flags |= EligibleAttester
	}
	//// TODO: also consider slashed but non-withdrawn validators?
	//for i := 0; i < count; i++ {
	//	v := state.Validators[i]
	//	vStatus := &status.ValidatorStatuses[i]
	//	if v.Slashed && currentEpoch < v.WithdrawableEpoch {
	//		vStatus.Flags |= EligibleAttester
	//	}
	//}
	prevTargetStake := Gwei(0)
	for i := range out {
		if out[i].Flags.HasMarkers(EligibleAttester | PrevEpochBoundaryAttester) {
			prevTargetStake += meta.EffectiveBalance(ValidatorIndex(i))
		}
	}
	return
}
