package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type AttesterStatuses []AttesterStatus

func (ats AttesterStatuses) GetAttesterStatus(index ValidatorIndex) AttesterStatus {
	return ats[index]
}

type AttesterStatusFeature struct {
	State *AttestationsState
	Meta interface {
		meta.VersioningMeta
		meta.RegistrySizeMeta
		meta.CrosslinkTimingMeta
		meta.CommitteeCountMeta
		meta.CrosslinkCommitteeMeta
		meta.EffectiveBalanceMeta
		meta.HistoryMeta
		meta.SlashedMeta
		meta.ActiveIndicesMeta
	}
}

func (f *AttesterStatusFeature) LoadAttesterStatuses() (out AttesterStatuses) {
	count := f.Meta.ValidatorCount()
	out = make(AttesterStatuses, count, count)

	currentEpoch := f.Meta.CurrentEpoch()
	prevEpoch := f.Meta.PreviousEpoch()
	previousBoundaryBlockRoot := f.Meta.GetBlockRootAtSlot(prevEpoch.GetStartSlot())
	//currentBoundaryBlockRoot := meta.GetBlockRootAtSlot(currentEpoch.GetStartSlot())

	participants := make([]ValidatorIndex, 0, MAX_VALIDATORS_PER_COMMITTEE)
	for _, att := range f.State.PreviousEpochAttestations {
		attBlockRoot := f.Meta.GetBlockRootAtSlot(att.Data.GetAttestationSlot(f.Meta))

		// attestation-target is already known to be previous-epoch, get it from the pre-computed shuffling directly.
		committee := f.Meta.GetCrosslinkCommittee(prevEpoch, att.Data.Crosslink.Shard)

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

			if !f.Meta.IsSlashed(p) {
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
	for _, index := range f.Meta.GetActiveValidatorIndices(currentEpoch) {
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
			prevTargetStake += f.Meta.EffectiveBalance(ValidatorIndex(i))
		}
	}
	return
}
