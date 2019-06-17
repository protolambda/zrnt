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

type ffg struct {
	sourceEpoch Epoch
	sourceRoot  Root
	targetEpoch Epoch
}

func ProcessAttestation(state *BeaconState, attestation *Attestation) error {
	data := &attestation.Data
	if data.Crosslink.Shard >= SHARD_COUNT {
		return errors.New("attestation data is invalid, shard out of range")
	}
	currentEpoch := state.Epoch()
	previousEpoch := state.PreviousEpoch()

	if data.TargetEpoch < previousEpoch {
		return errors.New("attestation data is invalid, target is too far in past")
	} else if data.TargetEpoch > currentEpoch {
		return errors.New("attestation data is invalid, target is in future")
	}

	attestationSlot := state.GetAttestationSlot(data)
	if !(state.Slot <= attestationSlot+SLOTS_PER_EPOCH) {
		return errors.New("attestation slot is too old")
	}
	if !(attestationSlot+MIN_ATTESTATION_INCLUSION_DELAY <= state.Slot) {
		return errors.New("attestation is too new")
	}

	var ffgData ffg
	var parentCrosslink *Crosslink

	if data.TargetEpoch == currentEpoch {
		ffgData = ffg{state.CurrentJustifiedEpoch, state.CurrentJustifiedRoot, currentEpoch}
		parentCrosslink = &state.CurrentCrosslinks[data.Crosslink.Shard]
	} else {
		ffgData = ffg{state.PreviousJustifiedEpoch, state.PreviousJustifiedRoot, previousEpoch}
		parentCrosslink = &state.PreviousCrosslinks[data.Crosslink.Shard]
	}

	// Check FFG data, crosslink data, and signature
	// -------------------------------------------------
	// FFG
	if ffgData != (ffg{data.SourceEpoch, data.SourceRoot, data.TargetEpoch}) {
		return errors.New("attestation data source fields are invalid")
	}
	if data.Crosslink.ParentRoot != ssz.HashTreeRoot(parentCrosslink, CrosslinkSSZ) {
		return errors.New("attestation parent crosslink is invalid")
	}

	// crosslink data
	if data.Crosslink.StartEpoch != parentCrosslink.EndEpoch {
		return fmt.Errorf("attestation start epoch is invalid,"+
			" does not match parent crosslink end: %d <> %d", data.Crosslink.StartEpoch, parentCrosslink.EndEpoch)
	}
	if parentEnd := parentCrosslink.EndEpoch + MAX_EPOCHS_PER_CROSSLINK; parentEnd < data.TargetEpoch {
		if data.Crosslink.EndEpoch != parentEnd {
			return fmt.Errorf("attestation end epoch is invalid,"+
				" does not match (parent crosslink end epoch + epochs per link): %d <> %d",
				data.Crosslink.EndEpoch, parentEnd)
		}
	} else {
		if data.Crosslink.EndEpoch != data.TargetEpoch {
			return fmt.Errorf("attestation end epoch is invalid,"+
				" does not match parent target epoch: %d <> %d", data.Crosslink.EndEpoch, data.TargetEpoch)
		}
	}
	if data.Crosslink.DataRoot != (Root{}) { //  # [to be removed in phase 1]
		return errors.New("attestation cannot reference a crosslink root yet, processing as phase 0")
	}

	// Check signature and bitfields
	if indexedAtt, err := state.ConvertToIndexed(attestation); err != nil {
		return fmt.Errorf("attestation could not be converted to an indexed attestation: %v", err)
	} else if err := state.ValidateIndexedAttestation(indexedAtt); err != nil {
		return fmt.Errorf("attestation could not be verified in its indexed form: %v", err)
	}

	// Cache pending attestation
	pendingAttestation := &PendingAttestation{
		Data:                *data,
		AggregationBitfield: attestation.AggregationBitfield,
		InclusionDelay:      state.Slot - attestationSlot,
		ProposerIndex:       state.GetBeaconProposerIndex(),
	}
	if data.TargetEpoch == currentEpoch {
		state.CurrentEpochAttestations = append(state.CurrentEpochAttestations, pendingAttestation)
	} else {
		state.PreviousEpochAttestations = append(state.PreviousEpochAttestations, pendingAttestation)
	}
	return nil
}
