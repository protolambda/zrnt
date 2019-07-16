package operations

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon/components"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type Attestations []Attestation

func (_ *Attestations) Limit() uint32 {
	return MAX_ATTESTATIONS
}

func (ops Attestations) Process(state *BeaconState) error {
	for _, op := range ops {
		if err := op.Process(state); err != nil {
			return err
		}
	}
	return nil
}

type Attestation struct {
	AggregationBits CommitteeBits
	Data AttestationData
	CustodyBits CommitteeBits
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

	if data.Target.Epoch < previousEpoch {
		return errors.New("attestation data is invalid, target is too far in past")
	} else if data.Target.Epoch > currentEpoch {
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

	if data.Target.Epoch == currentEpoch {
		if data.Source.Epoch != state.CurrentJustifiedCheckpoint.Epoch {
			return errors.New("attestation source epoch does not match current justified checkpoint")
		}
		parentCrosslink = &state.CurrentCrosslinks[data.Crosslink.Shard]
	} else {
		if data.Source.Epoch != state.PreviousJustifiedCheckpoint.Epoch {
			return errors.New("attestation source epoch does not match previous justified checkpoint")
		}
		parentCrosslink = &state.PreviousCrosslinks[data.Crosslink.Shard]
	}

	// Check crosslink against expected parent crosslink
	if data.Crosslink.ParentRoot != ssz.HashTreeRoot(parentCrosslink, CrosslinkSSZ) {
		return errors.New("attestation parent crosslink is invalid")
	}

	// crosslink data
	if data.Crosslink.StartEpoch != parentCrosslink.EndEpoch {
		return fmt.Errorf("attestation start epoch is invalid,"+
			" does not match parent crosslink end: %d <> %d", data.Crosslink.StartEpoch, parentCrosslink.EndEpoch)
	}
	if parentEnd := parentCrosslink.EndEpoch + MAX_EPOCHS_PER_CROSSLINK; parentEnd < data.Target.Epoch {
		if data.Crosslink.EndEpoch != parentEnd {
			return fmt.Errorf("attestation end epoch is invalid,"+
				" does not match (parent crosslink end epoch + epochs per link): %d <> %d",
				data.Crosslink.EndEpoch, parentEnd)
		}
	} else {
		if data.Crosslink.EndEpoch != data.Target.Epoch {
			return fmt.Errorf("attestation end epoch is invalid,"+
				" does not match parent target epoch: %d <> %d", data.Crosslink.EndEpoch, data.Target.Epoch)
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
		AggregationBits: attestation.AggregationBits,
		InclusionDelay:      state.Slot - attestationSlot,
		ProposerIndex:       state.GetBeaconProposerIndex(),
	}
	if data.Target.Epoch == currentEpoch {
		state.CurrentEpochAttestations = append(state.CurrentEpochAttestations, pendingAttestation)
	} else {
		state.PreviousEpochAttestations = append(state.PreviousEpochAttestations, pendingAttestation)
	}
	return nil
}

// Convert attestation to (almost) indexed-verifiable form
func (attestation *Attestation) ConvertToIndexed(state *BeaconState) (*IndexedAttestation, error) {
	bitLen := attestation.AggregationBits.BitLen()
	if custodyBitLen := attestation.CustodyBits.BitLen(); bitLen != custodyBitLen {
		return nil, fmt.Errorf("aggregation bits does not match custody size: %d <> %d", bitLen, custodyBitLen)
	}

	committee := state.PrecomputedData.GetCrosslinkCommittee(
		attestation.Data.Target.Epoch,
		attestation.Data.Crosslink.Shard,
	)
	if uint32(len(committee)) != bitLen {
		return nil, fmt.Errorf("committee size does not match bits size: %d <> %d", len(committee), bitLen)
	}

	bit1s := make([]ValidatorIndex, 0, len(committee))
	bit0s := make([]ValidatorIndex, 0, len(committee))
	for i := uint32(0); i < bitLen; i++ {
		if attestation.AggregationBits.GetBit(i) {
			if attestation.CustodyBits.GetBit(i) {
				bit1s = append(bit1s, committee[i])
			} else {
				bit0s = append(bit0s, committee[i])
			}
		} else {
			if attestation.CustodyBits.GetBit(i) {
				return nil, fmt.Errorf("custody bits not a subset of aggregations bits, different at: %d", i)
			}
		}
	}

	return &IndexedAttestation{
		CustodyBit0Indices: bit0s,
		CustodyBit1Indices: bit1s,
		Data:               attestation.Data,
		Signature:          attestation.Signature,
	}, nil
}
