package attestations

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
	"sort"
)

type AttestationProcessor interface {
	ProcessAttestations(ops []Attestation) error
	ProcessAttestation(attestation *Attestation) error
}

type AttestationFeature struct {
	State AttestationsProps
	Meta  interface {
		meta.Versioning
		meta.BeaconCommittees
		meta.CommitteeCount
		meta.Finality
		meta.RegistrySize
		meta.Pubkeys
		meta.SigDomain
		meta.Proposers
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

var AttestationSSZ = zssz.GetSSZ((*Attestation)(nil))

type Attestation struct {
	AggregationBits CommitteeBits
	Data            AttestationData
	Signature       BLSSignature
}

var AttestationType = &ContainerType{
	{"aggregation_bits", CommitteeBitsType},
	{"data", AttestationDataType},
	{"signature", BLSSignatureType},
}

func (f *AttestationFeature) ProcessAttestation(attestation *Attestation) error {
	data := &attestation.Data

	// Check slot
	currentSlot, err := f.Meta.CurrentSlot()
	if err != nil {
		return err
	}
	if !(currentSlot <= data.Slot+SLOTS_PER_EPOCH) {
		return errors.New("attestation slot is too old")
	}
	if !(data.Slot+MIN_ATTESTATION_INCLUSION_DELAY <= currentSlot) {
		return errors.New("attestation is too new")
	}

	currentEpoch := currentSlot.ToEpoch()
	previousEpoch := currentEpoch.Previous()

	// Check target
	if data.Target.Epoch < previousEpoch {
		return errors.New("attestation data is invalid, target is too far in past")
	} else if data.Target.Epoch > currentEpoch {
		return errors.New("attestation data is invalid, target is in future")
	}
	// And if it matches the slot
	if data.Target.Epoch != data.Slot.ToEpoch() {
		return errors.New("attestation data is invalid, slot epoch does not match target epoch")
	}

	// Check committee index
	if commCount, err := f.Meta.GetCommitteeCountAtSlot(data.Slot); err != nil {
		return err
	} else if uint64(data.Index) >= commCount {
		return errors.New("attestation data is invalid, committee index out of range")
	}

	// Check source
	if data.Target.Epoch == currentEpoch {
		if currentJustified, err := f.Meta.CurrentJustified(); err != nil {
			return err
		} else if data.Source != currentJustified {
			return errors.New("attestation source does not match current justified checkpoint")
		}
	} else {
		if previousJustified, err := f.Meta.PreviousJustified(); err != nil {
			return err
		} else if data.Source != previousJustified {
			return errors.New("attestation source does not match previous justified checkpoint")
		}
	}

	// Check signature and bitfields
	committee, err := f.Meta.GetBeaconCommittee(data.Slot, data.Index)
	if err != nil {
		return err
	}
	if indexedAtt, err := attestation.ConvertToIndexed(committee); err != nil {
		return fmt.Errorf("attestation could not be converted to an indexed attestation: %v", err)
	} else if err := indexedAtt.Validate(f.Meta); err != nil {
		return fmt.Errorf("attestation could not be verified in its indexed form: %v", err)
	}

	// TODO
	//proposerIndex, err := f.Meta.GetBeaconProposerIndex(currentSlot)
	//if err != nil {
	//	return err
	//}
	//// Cache pending attestation
	//pendingAttestation := &PendingAttestation{
	//	Data:            *data,
	//	AggregationBits: attestation.AggregationBits,
	//	InclusionDelay:  currentSlot - data.Slot,
	//	ProposerIndex:   proposerIndex,
	//}
	//if data.Target.Epoch == currentEpoch {
	//	f.State.CurrentEpochAttestations = append(f.State.CurrentEpochAttestations, pendingAttestation)
	//} else {
	//	f.State.PreviousEpochAttestations = append(f.State.PreviousEpochAttestations, pendingAttestation)
	//}
	return nil
}

// Convert attestation to (almost) indexed-verifiable form
func (attestation *Attestation) ConvertToIndexed(committee []ValidatorIndex) (*IndexedAttestation, error) {
	bitLen := attestation.AggregationBits.BitLen()
	if uint64(len(committee)) != bitLen {
		return nil, fmt.Errorf("committee size does not match bits size: %d <> %d", len(committee), bitLen)
	}

	participants := make([]ValidatorIndex, 0, len(committee))
	for i := uint64(0); i < bitLen; i++ {
		if attestation.AggregationBits.GetBit(i) {
			participants = append(participants, committee[i])
		}
	}
	sort.Slice(participants, func(i int, j int) bool {
		return participants[i] < participants[j]
	})

	return &IndexedAttestation{
		AttestingIndices: participants,
		Data:             attestation.Data,
		Signature:        attestation.Signature,
	}, nil
}
