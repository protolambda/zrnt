package operations

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bitfield"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)


type Attestation struct {
	// Attester aggregation bitfield
	AggregationBitfield bitfield.Bitfield
	// Attestation data
	Data AttestationData
	// Custody bitfield
	CustodyBitfield bitfield.Bitfield
	// BLS aggregate signature
	Signature BLSSignature
}

type ffg struct {
	sourceEpoch Epoch
	sourceRoot  Root
	targetEpoch Epoch
}

func (attestation *Attestation) Process(state *BeaconState) error {
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
	if indexedAtt, err := attestation.ConvertToIndexed(state); err != nil {
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


// Convert attestation to (almost) indexed-verifiable form
func (attestation *Attestation) ConvertToIndexed(state *BeaconState) (*IndexedAttestation, error) {
	if a, b := len(attestation.AggregationBitfield), len(attestation.CustodyBitfield); a != b {
		return nil, fmt.Errorf("aggregation bitfield does not match custody bitfield size: %d <> %d", a, b)
	}
	participants, err := state.GetAttestingIndices(&attestation.Data, &attestation.AggregationBitfield)
	if err != nil {
		return nil, errors.New("participants could not be derived from aggregation_bitfield")
	}
	custodyBit1Indices, err := state.GetAttestingIndices(&attestation.Data, &attestation.CustodyBitfield)
	if err != nil {
		return nil, errors.New("participants could not be derived from custody_bitfield")
	}
	if len(custodyBit1Indices) > len(participants) {
		return nil, fmt.Errorf("attestation has more custody bits set (%d) than participants allowed (%d)",
			len(custodyBit1Indices), len(participants))
	}
	// everyone who is a participant, and has not a custody bit set to 1, is part of the 0 custody bit indices.
	custodyBit0Indices := make([]ValidatorIndex, 0, len(participants)-len(custodyBit1Indices))
	participants.ZigZagJoin(custodyBit1Indices, nil, func(i ValidatorIndex) {
		custodyBit0Indices = append(custodyBit0Indices, i)
	})
	return &IndexedAttestation{
		CustodyBit0Indices: custodyBit0Indices,
		CustodyBit1Indices: custodyBit1Indices,
		Data:               attestation.Data,
		Signature:          attestation.Signature,
	}, nil
}
