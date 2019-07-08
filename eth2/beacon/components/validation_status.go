package components

import (
	. "github.com/protolambda/zrnt/eth2/core"
)

type ValidatorStatusFlag uint64

func (flags ValidatorStatusFlag) hasMarkers(markers ValidatorStatusFlag) bool {
	return flags&markers == markers
}

const (
	PrevEpochAttester ValidatorStatusFlag = 1 << iota
	MatchingHeadAttester
	EpochBoundaryAttester
	UnslashedAttester
	EligibleAttester
)

type ValidatorStatus struct {
	// no delay (i.e. 0) by default
	InclusionDelay Slot
	Proposer       ValidatorIndex
	Flags          ValidatorStatusFlag
}

type ValidationStatus []ValidatorStatus

func (state *BeaconState) ValidationStatus() (vsl ValidationStatus) {
	vsl = make(ValidationStatus, len(state.Validators), len(state.Validators))

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

			if !state.Validators[p].Slashed {
				status.Flags |= UnslashedAttester
			}

			// remember the participant as one of the good validators
			status.Flags |= PrevEpochAttester

			// If the attestation is for the boundary:
			if att.Data.TargetRoot == previousBoundaryBlockRoot {
				status.Flags |= EpochBoundaryAttester
			}
			// If the attestation is for the head (att the time of attestation):
			if att.Data.BeaconBlockRoot == attBlockRoot {
				status.Flags |= MatchingHeadAttester
			}
		}
	}
	currentEpoch := state.Epoch()
	for i := 0; i < len(vsl); i++ {
		v := state.Validators[i]
		status := &vsl[i]
		if v.IsActive(currentEpoch) || (v.Slashed && currentEpoch < v.WithdrawableEpoch) {
			status.Flags |= EligibleAttester
		}
	}
	return
}
