package beacon

import (
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
)

type SignedAggregateAndProof struct {
	Message   AggregateAndProof `json:"message"`
	Signature BLSSignature      `json:"signature"`
}

func (a *SignedAggregateAndProof) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&a.Message), &a.Signature)
}

func (a *SignedAggregateAndProof) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&a.Message), &a.Signature)
}

func (a *SignedAggregateAndProof) ByteLength(spec *Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&a.Message), &a.Signature)
}

func (a *SignedAggregateAndProof) FixedLength(*Spec) uint64 {
	return 0
}

func (a *SignedAggregateAndProof) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(spec.Wrap(&a.Message), &a.Signature)
}

type AggregateAndProof struct {
	AggregatorIndex ValidatorIndex `json:"aggregator_index"`
	Aggregate       Attestation    `json:"aggregate"`
	SelectionProof  BLSSignature   `json:"selection_proof"`
}

func (a *AggregateAndProof) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Container(&a.AggregatorIndex, spec.Wrap(&a.Aggregate), &a.SelectionProof)
}

func (a *AggregateAndProof) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Container(&a.AggregatorIndex, spec.Wrap(&a.Aggregate), &a.SelectionProof)
}

func (a *AggregateAndProof) ByteLength(spec *Spec) uint64 {
	return codec.ContainerLength(&a.AggregatorIndex, spec.Wrap(&a.Aggregate), &a.SelectionProof)
}

func (a *AggregateAndProof) FixedLength(*Spec) uint64 {
	return 0
}

func (a *AggregateAndProof) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&a.AggregatorIndex, spec.Wrap(&a.Aggregate), &a.SelectionProof)
}
