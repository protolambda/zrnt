package block_processing

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessBlockAttestations(state *BeaconState, block *BeaconBlock) error {
	if len(block.Body.Attestations) > MAX_ATTESTATIONS {
		return errors.New("too many attestations")
	}
	for _, attestation := range block.Body.Attestations {
		if err := ProcessAttestation(state, &attestation); err != nil {
			return err
		}
	}
	return nil
}

func ProcessAttestation(state *BeaconState, attestation *Attestation) error {
	data := &attestation.Data
	if data.Shard >= SHARD_COUNT {
		return errors.New("attestation data is invalid, shard out of range")
	}
	attestationSlot := state.GetAttestationSlot(data)
	if !(state.Slot <= attestationSlot+SLOTS_PER_EPOCH) {
		return errors.New("attestation slot is too old")
	}
	if !(attestationSlot+MIN_ATTESTATION_INCLUSION_DELAY <= state.Slot) {
		return errors.New("attestation is too new")
	}
	// Check target epoch, source epoch, and source crosslink
	targetEpoch := data.TargetEpoch
	sourceEpoch := data.SourceEpoch
	sourceRoot := data.SourceRoot
	sourceCrosslink := data.PreviousCrosslinkRoot
	if !((targetEpoch == state.Epoch() &&
		sourceEpoch == state.CurrentJustifiedEpoch &&
		sourceRoot == state.CurrentJustifiedRoot &&
		sourceCrosslink == ssz.HashTreeRoot(state.CurrentCrosslinks[data.Shard], CrosslinkSSZ)) ||
		(targetEpoch == state.PreviousEpoch() &&
			sourceEpoch == state.PreviousJustifiedEpoch &&
			sourceRoot == state.PreviousJustifiedRoot) &&
			sourceCrosslink == ssz.HashTreeRoot(state.PreviousCrosslinks[data.Shard], CrosslinkSSZ)) {
		return errors.New("attestation does not match recent state justification")
	}

	// Check crosslink data
	if attestation.Data.CrosslinkDataRoot != (Root{}) { //  # [to be removed in phase 1]
		return errors.New("attestation cannot reference a crosslink root yet, processing as phase 0")
	}

	// Check signature and bitfields
	if indexedAtt, err := state.ConvertToIndexed(attestation); err != nil {
		return fmt.Errorf("attestation could not be converted to an indexed attestation: %v", err)
	} else if err := state.VerifyIndexedAttestation(indexedAtt); err != nil {
		return fmt.Errorf("attestation could not be verified in its indexed form: %v", err)
	}

	// Cache pending attestation
	pendingAttestation := &PendingAttestation{
		Data:                *data,
		AggregationBitfield: attestation.AggregationBitfield,
		InclusionDelay:      state.Slot - attestationSlot,
		ProposerIndex:       state.GetBeaconProposerIndex(),
	}
	if targetEpoch == state.Epoch() {
		state.CurrentEpochAttestations = append(state.CurrentEpochAttestations, pendingAttestation)
	} else {
		state.PreviousEpochAttestations = append(state.PreviousEpochAttestations, pendingAttestation)
	}
	return nil
}
