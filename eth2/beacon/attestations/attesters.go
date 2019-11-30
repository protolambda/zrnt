package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type AttesterStatusFeature struct {
	State *AttestationsProps
	Meta  interface {
		meta.Versioning
		meta.RegistrySize
		meta.CommitteeCount
		meta.BeaconCommittees
		meta.History
		meta.SlashedIndices
		meta.ActiveIndices
		meta.ValidatorEpochData
	}
}

func (f *AttesterStatusFeature) GetAttesterStatuses() (out []AttesterStatus, err error) {
	count, err := f.Meta.ValidatorCount()
	if err != nil {
		return nil, err
	}

	currentEpoch, err := f.Meta.CurrentEpoch()
	if err != nil {
		return nil, err
	}
	prevEpoch, err := f.Meta.PreviousEpoch()
	if err != nil {
		return nil, err
	}

	out = make([]AttesterStatus, count, count)

	for i := ValidatorIndex(0); i < ValidatorIndex(len(out)); i++ {
		status := &out[i]
		slashed, err := f.Meta.IsSlashed(i)
		if err != nil {
			return nil, err
		}
		if !slashed {
			status.Flags |= UnslashedAttester
		}
		if active, err := f.Meta.IsActive(i, currentEpoch); err != nil {
			return nil, err
		} else if active {
			status.Flags |= EligibleAttester
		} else if slashed {
			withdrawableEpoch, err := f.Meta.WithdrawableEpoch(i)
			if err != nil {
				return nil, err
			} else if prevEpoch+1 < withdrawableEpoch {
				status.Flags |= EligibleAttester
			}
		}
		status.AttestedProposer = ValidatorIndexMarker
	}

	processEpoch := func(
		attestations []*PendingAttestation, epoch Epoch,
		sourceFlag, targetFlag, headFlag AttesterFlag) error {

		targetBlockRoot, err := f.Meta.GetBlockRootAtSlot(epoch.GetStartSlot())
		if err != nil {
			return err
		}
		participants := make([]ValidatorIndex, 0, MAX_VALIDATORS_PER_COMMITTEE)
		for _, att := range attestations {
			attBlockRoot, err := f.Meta.GetBlockRootAtSlot(att.Data.Slot)
			if err != nil {
				return err
			}

			// attestation-target is already known to be this epoch, get it from the pre-computed shuffling directly.
			committee, err := f.Meta.GetBeaconCommittee(att.Data.Slot, att.Data.Index)
			if err != nil {
				return err
			}

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
		return nil
	}
	prevPending, err := f.State.PreviousEpochAttestations.PendingAttestations()
	if err != nil {
		return nil, err
	}
	prevPendingRawAtts, err := prevPending.CollectRawPendingAttestations()
	if err != nil {
		return nil, err
	}
	if err := processEpoch(prevPendingRawAtts, prevEpoch,
		PrevSourceAttester, PrevTargetAttester, PrevHeadAttester); err != nil {
			return nil, err
	}
	currPending, err := f.State.CurrentEpochAttestations.PendingAttestations()
	if err != nil {
		return nil, err
	}
	currPendingRawAtts, err := currPending.CollectRawPendingAttestations()
	if err != nil {
		return nil, err
	}
	if err := processEpoch(currPendingRawAtts, currentEpoch,
		CurrSourceAttester, CurrTargetAttester, CurrHeadAttester); err != nil {
			return nil, err
	}
	return
}
