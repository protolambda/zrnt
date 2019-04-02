package block_processing

import (
	"errors"
	"github.com/protolambda/zrnt/eth2/beacon"
)

func ProcessBlockAttestations(state *beacon.BeaconState, block *beacon.BeaconBlock) error {
	if len(block.Body.Attestations) > beacon.MAX_ATTESTATIONS {
		return errors.New("too many attestations")
	}
	for _, attestation := range block.Body.Attestations {
		if err := ProcessAttestation(state, &attestation); err != nil {
			return err
		}
	}
	return nil
}

func ProcessAttestation(state *beacon.BeaconState, attestation *beacon.Attestation) error {

	if !(beacon.GENESIS_SLOT <= attestation.Data.Slot &&
		state.Slot-beacon.SLOTS_PER_EPOCH <= attestation.Data.Slot) {
		return errors.New("attestation slot is too old")
	}
	if !(attestation.Data.Slot <= state.Slot-beacon.MIN_ATTESTATION_INCLUSION_DELAY) {
		return errors.New("attestation is too new")
	}
	// Check target epoch, source epoch, and source root
	targetEpoch := attestation.Data.Slot.ToEpoch()
	sourceEpoch := attestation.Data.SourceEpoch
	sourceRoot := attestation.Data.SourceRoot
	if !((targetEpoch == state.Epoch() && sourceEpoch == state.CurrentJustifiedEpoch && sourceRoot == state.CurrentJustifiedRoot) ||
		(targetEpoch == state.PreviousEpoch() && sourceEpoch == state.PreviousJustifiedEpoch && sourceRoot == state.PreviousJustifiedRoot)) {
		return errors.New("attestation does not match recent state justification")
	}

	// Check crosslink data
	if attestation.Data.CrosslinkDataRoot == (beacon.Root{}) { //  # [to be removed in phase 1]
		return errors.New("attestation cannot reference a crosslink root yet, processing as phase 0")
	}
	if !(
	// Case 1: latest crosslink matches previous crosslink
		state.LatestCrosslinks[attestation.Data.Shard] == attestation.Data.PreviousCrosslink ||
		// Case 2: latest crosslink matches current crosslink
			state.LatestCrosslinks[attestation.Data.Shard] == beacon.Crosslink{CrosslinkDataRoot: attestation.Data.CrosslinkDataRoot, Epoch: attestation.Data.Slot.ToEpoch()}) {
		return errors.New("attestation crosslinking invalid")
	}

	// Check signature and bitfields
	if indexedAtt, err := state.ConvertToIndexed(attestation); err != nil {
		return errors.New("attestation could not be converted to an indexed attestation")
	} else if !state.VerifyIndexedAttestation(indexedAtt) {
		return errors.New("attestation could not be verified in its indexed form")
	}

	// Cache pending attestation
	pendingAttestation := beacon.PendingAttestation{
		Data:                attestation.Data,
		AggregationBitfield: attestation.AggregationBitfield,
		CustodyBitfield:     attestation.CustodyBitfield,
		InclusionSlot:       state.Slot,
	}
	if targetEpoch == state.Epoch() {
		state.CurrentEpochAttestations = append(state.CurrentEpochAttestations, pendingAttestation)
	} else {
		state.PreviousEpochAttestations = append(state.CurrentEpochAttestations, pendingAttestation)
	}
	return nil
}
