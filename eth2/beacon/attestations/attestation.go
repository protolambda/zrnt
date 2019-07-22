package attestations

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type AttestationFeature struct {
	State *AttestationsState
	Meta  interface {
		meta.VersioningMeta
		meta.CrosslinkCommitteeMeta
		meta.CrosslinkTimingMeta
		meta.CommitteeCountMeta
		meta.CrosslinkMeta
		meta.FinalityMeta
		meta.RegistrySizeMeta
		meta.PubkeyMeta
		meta.ProposingMeta
	}
}

func (f *AttestationFeature) ProcessAttestations(ops []Attestation) error {
	for i := range ops {
		if err := f.ProcessAttestation(&ops[i]); err != nil {
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

func (f *AttestationFeature) ProcessAttestation(attestation *Attestation) error {
	data := &attestation.Data
	if data.Crosslink.Shard >= SHARD_COUNT {
		return errors.New("attestation data is invalid, shard out of range")
	}
	currentEpoch := f.Meta.CurrentEpoch()
	previousEpoch := f.Meta.PreviousEpoch()

	if data.Target.Epoch < previousEpoch {
		return errors.New("attestation data is invalid, target is too far in past")
	} else if data.Target.Epoch > currentEpoch {
		return errors.New("attestation data is invalid, target is in future")
	}

	currentSlot := f.Meta.CurrentSlot()
	attestationSlot := data.GetAttestationSlot(f.Meta)
	if !(currentSlot <= attestationSlot+SLOTS_PER_EPOCH) {
		return errors.New("attestation slot is too old")
	}
	if !(attestationSlot+MIN_ATTESTATION_INCLUSION_DELAY <= currentSlot) {
		return errors.New("attestation is too new")
	}

	var parentCrosslink *Crosslink

	if data.Target.Epoch == currentEpoch {
		if data.Source.Epoch != f.Meta.CurrentJustified().Epoch {
			return errors.New("attestation source epoch does not match current justified checkpoint")
		}
		parentCrosslink = f.Meta.GetCurrentCrosslink(data.Crosslink.Shard)
	} else {
		if data.Source.Epoch != f.Meta.PreviousJustified().Epoch {
			return errors.New("attestation source epoch does not match previous justified checkpoint")
		}
		parentCrosslink = f.Meta.GetPreviousCrosslink(data.Crosslink.Shard)
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
	committee := f.Meta.GetCrosslinkCommittee(
		attestation.Data.Target.Epoch,
		attestation.Data.Crosslink.Shard,
	)
	if indexedAtt, err := attestation.ConvertToIndexed(committee); err != nil {
		return fmt.Errorf("attestation could not be converted to an indexed attestation: %v", err)
	} else if err := indexedAtt.Validate(f.Meta); err != nil {
		return fmt.Errorf("attestation could not be verified in its indexed form: %v", err)
	}

	// Cache pending attestation
	pendingAttestation := &PendingAttestation{
		Data:            *data,
		AggregationBits: attestation.AggregationBits,
		InclusionDelay:  currentSlot - attestationSlot,
		ProposerIndex:   f.Meta.GetBeaconProposerIndex(currentSlot),
	}
	if data.Target.Epoch == currentEpoch {
		f.State.CurrentEpochAttestations = append(f.State.CurrentEpochAttestations, pendingAttestation)
	} else {
		f.State.PreviousEpochAttestations = append(f.State.PreviousEpochAttestations, pendingAttestation)
	}
	return nil
}

// Convert attestation to (almost) indexed-verifiable form
func (attestation *Attestation) ConvertToIndexed(committee []ValidatorIndex) (*IndexedAttestation, error) {
	bitLen := attestation.AggregationBits.BitLen()
	if custodyBitLen := attestation.CustodyBits.BitLen(); bitLen != custodyBitLen {
		return nil, fmt.Errorf("aggregation bits does not match custody size: %d <> %d", bitLen, custodyBitLen)
	}

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
