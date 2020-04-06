package beacon

import (
	"errors"
	"fmt"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
	"sort"
)

var AttestationsType = ListType(AttestationType, MAX_ATTESTATIONS)

var AttestationType = ContainerType("Attestation", []FieldDef{
	{"aggregation_bits", CommitteeBitsType},
	{"data", AttestationDataType},
	{"signature", BLSSignatureType},
})

var AttestationSSZ = zssz.GetSSZ((*Attestation)(nil))

type Attestation struct {
	AggregationBits CommitteeBits
	Data            AttestationData
	Signature       BLSSignature
}

type Attestations []Attestation

func (*Attestations) Limit() uint64 {
	return MAX_ATTESTATIONS
}

func (state *BeaconStateView) ProcessAttestations(epc *EpochsContext, ops []Attestation) error {
	for i := range ops {
		if err := state.ProcessAttestation(epc, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

func (state *BeaconStateView) ProcessAttestation(epc *EpochsContext, attestation *Attestation) error {
	data := &attestation.Data

	// Check slot
	currentSlot, err := state.Slot()
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
	if commCount, err := epc.GetCommitteeCountAtSlot(data.Slot); err != nil {
		return err
	} else if uint64(data.Index) >= commCount {
		return errors.New("attestation data is invalid, committee index out of range")
	}

	// Check source
	if data.Target.Epoch == currentEpoch {
		currentJustified, err := state.CurrentJustifiedCheckpoint()
		if err != nil {
			return err
		}
		currJustRaw, err := currentJustified.Raw()
		if err != nil {
			return err
		}
		if data.Source != currJustRaw {
			return errors.New("attestation source does not match current justified checkpoint")
		}
	} else {
		previousJustified, err := state.PreviousJustifiedCheckpoint()
		if err != nil {
			return err
		}
		prevJustRaw, err := previousJustified.Raw()
		if err != nil {
			return err
		}
		if data.Source != prevJustRaw {
			return errors.New("attestation source does not match previous justified checkpoint")
		}
	}

	// Check signature and bitfields
	committee, err := epc.GetBeaconCommittee(data.Slot, data.Index)
	if err != nil {
		return err
	}
	if indexedAtt, err := attestation.ConvertToIndexed(committee); err != nil {
		return fmt.Errorf("attestation could not be converted to an indexed attestation: %v", err)
	} else if err := state.ValidateIndexedAttestation(epc, indexedAtt); err != nil {
		return fmt.Errorf("attestation could not be verified in its indexed form: %v", err)
	}

	// TODO pending attestation to att node, append to tree
	proposerIndex, err := epc.GetBeaconProposer(currentSlot)
	if err != nil {
		return err
	}
	// Cache pending attestation
	pendingAttestation := PendingAttestation{
		Data:            *data,
		AggregationBits: attestation.AggregationBits,
		InclusionDelay:  currentSlot - data.Slot,
		ProposerIndex:   proposerIndex,
	}.View()

	if data.Target.Epoch == currentEpoch {
		atts, err := state.CurrentEpochAttestations()
		if err != nil {
			return err
		}
		if err := atts.Append(pendingAttestation); err != nil {
			return err
		}
	} else {
		atts, err := state.PreviousEpochAttestations()
		if err != nil {
			return err
		}
		if err := atts.Append(pendingAttestation); err != nil {
			return err
		}
	}
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
