package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var ShardBlobReferenceType = ContainerType("ShardBlobReference", []FieldDef{
	{"slot", common.SlotType},
	{"shard", common.ShardType},
	{"body_root", RootType},
	{"proposer_index", common.ValidatorIndexType},
})

type ShardBlobReference struct {
	// Slot that this header is intended for
	Slot common.Slot `json:"slot" yaml:"slot"`
	// Shard that this header is intended for
	Shard common.Shard `json:"shard" yaml:"shard"`

	// Hash-tree-root of ShardBlobBody
	BodyRoot common.Root `json:"body_root" yaml:"body_root"`

	// Proposer of the shard-blob
	ProposerIndex common.ValidatorIndex `json:"proposer_index" yaml:"proposer_index"`
}

func (v *ShardBlobReference) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.Slot, &v.Shard, &v.BodyRoot, &v.ProposerIndex)
}

func (v *ShardBlobReference) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.Slot, &v.Shard, &v.BodyRoot, &v.ProposerIndex)
}

func (v *ShardBlobReference) ByteLength() uint64 {
	return ShardBlobReferenceType.TypeByteLength()
}

func (*ShardBlobReference) FixedLength() uint64 {
	return ShardBlobReferenceType.TypeByteLength()
}

func (v *ShardBlobReference) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.Slot, &v.Shard, &v.BodyRoot, &v.ProposerIndex)
}

var SignedShardBlobReferenceType = ContainerType("SignedShardBlobReference", []FieldDef{
	{"message", ShardBlobReferenceType},
	{"signature", common.BLSSignatureType},
})

type SignedShardBlobReference struct {
	Message   ShardBlobReference  `json:"message" yaml:"message"`
	Signature common.BLSSignature `json:"signature" yaml:"signature"`
}

func (v *SignedShardBlobReference) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.Message, &v.Signature)
}

func (v *SignedShardBlobReference) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.Message, &v.Signature)
}

func (v *SignedShardBlobReference) ByteLength() uint64 {
	return SignedShardBlobReferenceType.TypeByteLength()
}

func (*SignedShardBlobReference) FixedLength() uint64 {
	return SignedShardBlobReferenceType.TypeByteLength()
}

func (v *SignedShardBlobReference) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.Message, v.Signature)
}
