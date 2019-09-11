package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type AttesterStatusFeature struct {
	State *AttestationsState
	Meta  interface {
		meta.Versioning
		meta.RegistrySize
		meta.CrosslinkTiming
		meta.CommitteeCount
		meta.CrosslinkCommittees
		meta.History
		meta.SlashedIndices
		meta.ActiveIndices
		meta.ValidatorEpochData
	}
}

func (f *AttesterStatusFeature) GetAttesterStatuses() (out []AttesterStatus) {
	count := f.Meta.ValidatorCount()

	currentEpoch := f.Meta.CurrentEpoch()
	prevEpoch := f.Meta.PreviousEpoch()

	out = make([]AttesterStatus, count, count)

	for i := ValidatorIndex(0); i < ValidatorIndex(len(out)); i++ {
		status := &out[i]
		if !f.Meta.IsSlashed(i) {
			status.Flags |= UnslashedAttester
		}
		if f.Meta.IsActive(i, currentEpoch) {
			status.Flags |= EligibleAttester
		} else if f.Meta.IsSlashed(i) && prevEpoch+1 < f.Meta.WithdrawableEpoch(i) {
			status.Flags |= EligibleAttester
		}
		status.AttestedProposer = ValidatorIndexMarker
	}

	processEpoch := func(
		attestations []*PendingAttestation, epoch Epoch,
		sourceFlag, targetFlag, headFlag AttesterFlag) {

		targetBlockRoot := f.Meta.GetBlockRootAtSlot(epoch.GetStartSlot())
		participants := make([]ValidatorIndex, 0, MAX_VALIDATORS_PER_COMMITTEE)
		for _, att := range attestations {
			attBlockRoot := f.Meta.GetBlockRootAtSlot(att.Data.GetAttestationSlot(f.Meta))

			// attestation-target is already known to be this epoch, get it from the pre-computed shuffling directly.
			committee := f.Meta.GetCrosslinkCommittee(epoch, att.Data.Crosslink.Shard)

			participants = participants[:0]                   // reset old slice (re-used in for loop)
			participants = append(participants, committee...) // add committee indices

			if epoch == prevEpoch {
				for _, p := range participants {
					status := &out[p]

					// If the attestation is the earliest, i.e. has the smallest delay
					if status.AttestedProposer == ValidatorIndexMarker || status.InclusionDelay > att.InclusionDelay {
						status.InclusionDelay = att.InclusionDelay
						status.AttestedProposer = att.ProposerIndex
					}
				}
			}

			participants = att.AggregationBits.FilterParticipants(participants) // only keep the participants
			for _, p := range participants {
				status := &out[p]

				// remember the participant as one of the good validators
				status.Flags |= sourceFlag

				// If the attestation is for the boundary:
				if att.Data.Target.Root == targetBlockRoot {
					status.Flags |= targetFlag
				}
				// If the attestation is for the head (att the time of attestation):
				if att.Data.BeaconBlockRoot == attBlockRoot {
					status.Flags |= headFlag
				}
			}
		}
	}
	processEpoch(f.State.PreviousEpochAttestations, prevEpoch,
		PrevSourceAttester, PrevTargetAttester, PrevHeadAttester)
	processEpoch(f.State.CurrentEpochAttestations, currentEpoch,
		CurrSourceAttester, CurrTargetAttester, CurrHeadAttester)
	return
}
