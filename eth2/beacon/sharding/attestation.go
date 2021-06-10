package sharding

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func BlockAttestationsType(spec *common.Spec) ListTypeDef {
	return ListType(AttestationType(spec), spec.MAX_ATTESTATIONS)
}

func AttestationType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("Attestation", []FieldDef{
		{"aggregation_bits", phase0.AttestationBitsType(spec)},
		{"data", AttestationDataType},
		{"signature", common.BLSSignatureType},
	})
}

type Attestation struct {
	AggregationBits phase0.AttestationBits     `json:"aggregation_bits" yaml:"aggregation_bits"`
	Data            AttestationData     `json:"data" yaml:"data"`
	Signature       common.BLSSignature `json:"signature" yaml:"signature"`
}

func (a *Attestation) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&a.AggregationBits), &a.Data, &a.Signature)
}

func (a *Attestation) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&a.AggregationBits), &a.Data, &a.Signature)
}

func (a *Attestation) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&a.AggregationBits), &a.Data, &a.Signature)
}

func (a *Attestation) FixedLength(*common.Spec) uint64 {
	return 0
}

func (a *Attestation) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&a.AggregationBits), &a.Data, a.Signature)
}

type Attestations []Attestation

func (a *Attestations) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, Attestation{})
		return spec.Wrap(&((*a)[i]))
	}, 0, spec.MAX_ATTESTATIONS)
}

func (a Attestations) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&a[i])
	}, 0, uint64(len(a)))
}

func (a Attestations) ByteLength(spec *common.Spec) (out uint64) {
	for _, v := range a {
		out += v.ByteLength(spec) + codec.OFFSET_SIZE
	}
	return
}

func (a *Attestations) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li Attestations) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&li[i])
		}
		return nil
	}, length, spec.MAX_ATTESTATIONS)
}

func ProcessAttestations(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, ops []Attestation) error {
	for i := range ops {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := ProcessAttestation(spec, epc, state, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

func ProcessAttestation(spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, attestation *Attestation) error {
	phase0Att := phase0.Attestation{
		AggregationBits: attestation.AggregationBits,
		Data:            phase0.AttestationData{
			Slot:            attestation.Data.Slot,
			Index:           attestation.Data.Index,
			BeaconBlockRoot: attestation.Data.BeaconBlockRoot,
			Source:          attestation.Data.Source,
			Target:          attestation.Data.Target,
		},
		Signature:       attestation.Signature,
	}
	if err := phase0.ProcessAttestation(spec, epc, state, &phase0Att); err != nil {
		return err
	}
	return UpdatePendingShardWork(spec, epc, state, attestation)
}

func UpdatePendingShardWork(spec *common.Spec, epc *common.EpochsContext, state *BeaconStateView, attestation *Attestation) error {
	// TODO
	return nil
}
