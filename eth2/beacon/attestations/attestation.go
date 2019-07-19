package attestations

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type AttestationReq interface {
	VersioningMeta
	CrosslinkCommitteeMeta
	CrosslinkMeta
	FinalityMeta
	RegistrySizeMeta
	PubkeyMeta
	ProposingMeta
}

func (state *AttestationsState) ProcessAttestations(meta AttestationReq, ops []Attestation) error {
	for i := range ops {
		if err := state.ProcessAttestation(meta, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

type Attestation struct {
	AggregationBits CommitteeBits
	Data            AttestationData
	CustodyBits     CommitteeBits
	Signature       BLSSignature
}

var crosslinkSSZ = zssz.GetSSZ((*Crosslink)(nil))

func (state *AttestationsState) ProcessAttestation(meta AttestationReq, attestation *Attestation) error {
	data := &attestation.Data
	if data.Crosslink.Shard >= SHARD_COUNT {
		return errors.New("attestation data is invalid, shard out of range")
	}
	currentEpoch := meta.Epoch()
	previousEpoch := meta.PreviousEpoch()

	if data.Target.Epoch < previousEpoch {
		return errors.New("attestation data is invalid, target is too far in past")
	} else if data.Target.Epoch > currentEpoch {
		return errors.New("attestation data is invalid, target is in future")
	}

	currentSlot := meta.Slot()
	attestationSlot := state.GetAttestationSlot(meta, data)
	if !(currentSlot <= attestationSlot+SLOTS_PER_EPOCH) {
		return errors.New("attestation slot is too old")
	}
	if !(attestationSlot+MIN_ATTESTATION_INCLUSION_DELAY <= currentSlot) {
		return errors.New("attestation is too new")
	}

	var parentCrosslink *Crosslink

	if data.Target.Epoch == currentEpoch {
		if data.Source.Epoch != meta.CurrentJustified().Epoch {
			return errors.New("attestation source epoch does not match current justified checkpoint")
		}
		parentCrosslink = meta.GetCurrentCrosslink(data.Crosslink.Shard)
	} else {
		if data.Source.Epoch != meta.PreviousJustified().Epoch {
			return errors.New("attestation source epoch does not match previous justified checkpoint")
		}
		parentCrosslink = meta.GetPreviousCrosslink(data.Crosslink.Shard)
	}

	// Check crosslink against expected parent crosslink
	if data.Crosslink.ParentRoot != ssz.HashTreeRoot(parentCrosslink, crosslinkSSZ) {
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
	if indexedAtt, err := attestation.ConvertToIndexed(meta); err != nil {
		return fmt.Errorf("attestation could not be converted to an indexed attestation: %v", err)
	} else if err := indexedAtt.Validate(meta); err != nil {
		return fmt.Errorf("attestation could not be verified in its indexed form: %v", err)
	}

	// Cache pending attestation
	pendingAttestation := &PendingAttestation{
		Data:            *data,
		AggregationBits: attestation.AggregationBits,
		InclusionDelay:  currentSlot - attestationSlot,
		ProposerIndex:   meta.GetBeaconProposerIndex(),
	}
	if data.Target.Epoch == currentEpoch {
		state.CurrentEpochAttestations = append(state.CurrentEpochAttestations, pendingAttestation)
	} else {
		state.PreviousEpochAttestations = append(state.PreviousEpochAttestations, pendingAttestation)
	}
	return nil
}

// Convert attestation to (almost) indexed-verifiable form
func (attestation *Attestation) ConvertToIndexed(meta CrosslinkCommitteeMeta) (*IndexedAttestation, error) {
	bitLen := attestation.AggregationBits.BitLen()
	if custodyBitLen := attestation.CustodyBits.BitLen(); bitLen != custodyBitLen {
		return nil, fmt.Errorf("aggregation bits does not match custody size: %d <> %d", bitLen, custodyBitLen)
	}

	committee := meta.GetCrosslinkCommittee(
		attestation.Data.Target.Epoch,
		attestation.Data.Crosslink.Shard,
	)
	if uint64(len(committee)) != bitLen {
		return nil, fmt.Errorf("committee size does not match bits size: %d <> %d", len(committee), bitLen)
	}

	bit1s := make([]ValidatorIndex, 0, len(committee))
	bit0s := make([]ValidatorIndex, 0, len(committee))
	for i := uint64(0); i < bitLen; i++ {
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
