package attestations

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type AttestersData struct {
	Epoch          Epoch
	Statuses       []AttesterStatus
	TotalStake     Gwei
	PrevTotalStake EpochStake
	CurrTotalStake EpochStake
}

func (atd AttestersData) GetAttesterStatus(index ValidatorIndex) AttesterStatus {
	return atd.Statuses[index]
}

func (atd AttestersData) GetTotalStake() Gwei {
	return atd.TotalStake
}

func (atd AttestersData) GetTotalEpochStake(epoch Epoch) EpochStake {
	if epoch == atd.Epoch {
		return atd.CurrTotalStake
	} else if epoch == atd.Epoch.Previous() {
		return atd.PrevTotalStake
	} else {
		panic("epoch out of computed range")
	}
}

type AttesterStatusFeature struct {
	State *AttestationsState
	Meta  interface {
		meta.Versioning
		meta.RegistrySize
		meta.CrosslinkTiming
		meta.CommitteeCount
		meta.CrosslinkCommittees
		meta.EffectiveBalances
		meta.History
		meta.SlashedIndices
		meta.ActiveIndices
		meta.ValidatorEpochData
	}
}

func (f *AttesterStatusFeature) LoadAttesterStatuses() (out *AttestersData) {
	count := f.Meta.ValidatorCount()

	currentEpoch := f.Meta.CurrentEpoch()
	prevEpoch := f.Meta.PreviousEpoch()

	out = &AttestersData{
		Epoch:    currentEpoch,
		Statuses: make([]AttesterStatus, count, count),
	}

	for i := ValidatorIndex(0); i < ValidatorIndex(len(out.Statuses)); i++ {
		status := &out.Statuses[i]
		if !f.Meta.IsSlashed(i) {
			status.Flags |= UnslashedAttester
		}
		if f.Meta.IsActive(i, currentEpoch) {
			out.TotalStake += f.Meta.EffectiveBalance(i)
			status.Flags |= EligibleAttester
		} else if f.Meta.IsSlashed(i) && prevEpoch+1 < f.Meta.WithdrawableEpoch(i) {
			status.Flags |= EligibleAttester
		}
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

			participants = participants[:0]                                     // reset old slice (re-used in for loop)
			participants = append(participants, committee...)                   // add committee indices
			participants = att.AggregationBits.FilterParticipants(participants) // only keep the participants
			for _, p := range participants {

				status := &out.Statuses[p]

				// If the attestation is the earliest, i.e. has the biggest delay
				if status.InclusionDelay < att.InclusionDelay {
					status.InclusionDelay = att.InclusionDelay
					status.AttestedProposer = att.ProposerIndex
				}

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

	for i := range out.Statuses {
		status := &out.Statuses[i]
		b := f.Meta.EffectiveBalance(ValidatorIndex(i))
		if status.Flags.HasMarkers(PrevSourceAttester | UnslashedAttester) {
			out.PrevTotalStake.SourceBalance += b
		}
		if status.Flags.HasMarkers(PrevTargetAttester | UnslashedAttester) {
			out.PrevTotalStake.TargetBalance += b
		}
		if status.Flags.HasMarkers(PrevHeadAttester | UnslashedAttester) {
			out.PrevTotalStake.HeadBalance += b
		}
		if status.Flags.HasMarkers(CurrSourceAttester | UnslashedAttester) {
			out.CurrTotalStake.SourceBalance += b
		}
		if status.Flags.HasMarkers(CurrTargetAttester | UnslashedAttester) {
			out.CurrTotalStake.TargetBalance += b
		}
		if status.Flags.HasMarkers(CurrHeadAttester | UnslashedAttester) {
			out.CurrTotalStake.HeadBalance += b
		}
	}
	return
}
