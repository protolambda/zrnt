package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var ShardBlobHeaderType = ContainerType("ShardBlobHeader", []FieldDef{
	{"slot", common.SlotType},
	{"shard", common.ShardType},
	{"body_summary", ShardBlobBodySummaryType},
	{"proposer_index", common.ValidatorIndexType},
})

type ShardBlobHeader struct {
	// Slot that this header is intended for
	Slot common.Slot `json:"slot" yaml:"slot"`
	// Shard that this header is intended for
	Shard common.Shard `json:"shard" yaml:"shard"`

	// SSZ-summary of ShardBlobBody
	BodySummary ShardBlobBodySummary `json:"body_summary" yaml:"body_summary"`

	// Proposer of the shard-blob
	ProposerIndex common.ValidatorIndex `json:"proposer_index" yaml:"proposer_index"`
}

func (v *ShardBlobHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.Slot, &v.Shard, &v.BodySummary, &v.ProposerIndex)
}

func (v *ShardBlobHeader) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.Slot, &v.Shard, &v.BodySummary, &v.ProposerIndex)
}

func (v *ShardBlobHeader) ByteLength() uint64 {
	return ShardBlobHeaderType.TypeByteLength()
}

func (*ShardBlobHeader) FixedLength() uint64 {
	return ShardBlobHeaderType.TypeByteLength()
}

func (v *ShardBlobHeader) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.Slot, &v.Shard, &v.BodySummary, &v.ProposerIndex)
}

var SignedShardBlobHeaderType = ContainerType("SignedShardBlobHeader", []FieldDef{
	{"message", ShardBlobHeaderType},
	{"signature", common.BLSSignatureType},
})

type SignedShardBlobHeader struct {
	Message   ShardBlobHeader     `json:"message" yaml:"message"`
	Signature common.BLSSignature `json:"signature" yaml:"signature"`
}

func (v *SignedShardBlobHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.Message, &v.Signature)
}

func (v *SignedShardBlobHeader) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.Message, &v.Signature)
}

func (v *SignedShardBlobHeader) ByteLength() uint64 {
	return SignedShardBlobHeaderType.TypeByteLength()
}

func (*SignedShardBlobHeader) FixedLength() uint64 {
	return SignedShardBlobHeaderType.TypeByteLength()
}

func (v *SignedShardBlobHeader) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.Message, v.Signature)
}

func BlockShardHeadersType(spec *common.Spec) ListTypeDef {
	return ListType(SignedShardBlobHeaderType, spec.MAX_SHARDS*spec.MAX_SHARD_HEADERS_PER_SHARD)
}

type ShardHeaders []SignedShardBlobHeader

func (a *ShardHeaders) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, SignedShardBlobHeader{})
		return &((*a)[i])
	}, SignedShardBlobHeaderType.TypeByteLength(), spec.MAX_SHARDS*spec.MAX_SHARD_HEADERS_PER_SHARD)
}

func (a ShardHeaders) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, SignedShardBlobHeaderType.TypeByteLength(), uint64(len(a)))
}

func (a ShardHeaders) ByteLength(*common.Spec) (out uint64) {
	return SignedShardBlobHeaderType.TypeByteLength() * uint64(len(a))
}

func (a *ShardHeaders) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li ShardHeaders) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, spec.MAX_SHARDS*spec.MAX_SHARD_HEADERS_PER_SHARD)
}
