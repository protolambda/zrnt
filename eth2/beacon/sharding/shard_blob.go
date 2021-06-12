package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func ShardBlobType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("ShardBlob", []FieldDef{
		{"slot", common.SlotType},
		{"shard", common.ShardType},
		{"body", ShardBlobBodyType(spec)},
		{"proposer_index", common.ValidatorIndexType},
	})
}

type ShardBlob struct {
	// Slot that this header is intended for
	Slot common.Slot `json:"slot" yaml:"slot"`
	// Shard that this header is intended for
	Shard common.Shard `json:"shard" yaml:"shard"`

	// Shard data with related commitments and beacon anchor
	Body ShardBlobBody `json:"body" yaml:"body"`

	// Proposer of the shard-blob
	ProposerIndex common.ValidatorIndex `json:"proposer_index" yaml:"proposer_index"`
}

func (b *ShardBlob) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&b.Slot, &b.Shard, spec.Wrap(&b.Body), &b.ProposerIndex)
}

func (b *ShardBlob) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&b.Slot, &b.Shard, spec.Wrap(&b.Body), &b.ProposerIndex)
}

func (b *ShardBlob) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&b.Slot, &b.Shard, spec.Wrap(&b.Body), &b.ProposerIndex)
}

func (*ShardBlob) FixedLength(spec *common.Spec) uint64 {
	return 0
}

func (b *ShardBlob) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&b.Slot, &b.Shard, spec.Wrap(&b.Body), &b.ProposerIndex)
}

func SignedShardBlobType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("SignedShardBlob", []FieldDef{
		{"message", ShardBlobType(spec)},
		{"signature", common.BLSSignatureType},
	})
}

type SignedShardBlob struct {
	Message   ShardBlob           `json:"message" yaml:"message"`
	Signature common.BLSSignature `json:"signature" yaml:"signature"`
}

func (b *SignedShardBlob) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedShardBlob) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.FixedLenContainer(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedShardBlob) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&b.Message), &b.Signature)
}

func (*SignedShardBlob) FixedLength(spec *common.Spec) uint64 {
	return 0
}

func (b *SignedShardBlob) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(spec.Wrap(&b.Message), b.Signature)
}
