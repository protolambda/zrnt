package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var ShardProposerSlashingType = ContainerType("ShardProposerSlashing", []FieldDef{
	{"signed_reference_1", SignedShardBlobReferenceType},
	{"signed_reference_2", SignedShardBlobReferenceType},
})

type ShardProposerSlashing struct {
	SignedReference1 SignedShardBlobReference `json:"signed_reference_1" yaml:"signed_reference_1"`
	SignedReference2 SignedShardBlobReference `json:"signed_reference_2" yaml:"signed_reference_2"`
}

func (v *ShardProposerSlashing) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.SignedReference1, &v.SignedReference2)
}

func (v *ShardProposerSlashing) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.SignedReference1, &v.SignedReference2)
}

func (v *ShardProposerSlashing) ByteLength() uint64 {
	return ShardProposerSlashingType.TypeByteLength()
}

func (*ShardProposerSlashing) FixedLength() uint64 {
	return ShardProposerSlashingType.TypeByteLength()
}

func (v *ShardProposerSlashing) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&v.SignedReference1, &v.SignedReference2)
}

func BlockShardProposerSlashingsType(spec *common.Spec) ListTypeDef {
	return ListType(ShardProposerSlashingType, spec.MAX_SHARD_PROPOSER_SLASHINGS)
}

type ShardProposerSlashings []ShardProposerSlashing

func (a *ShardProposerSlashings) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, ShardProposerSlashing{})
		return &((*a)[i])
	}, ShardProposerSlashingType.TypeByteLength(), spec.MAX_SHARD_PROPOSER_SLASHINGS)
}

func (a ShardProposerSlashings) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, ShardProposerSlashingType.TypeByteLength(), uint64(len(a)))
}

func (a ShardProposerSlashings) ByteLength(*common.Spec) (out uint64) {
	return ShardProposerSlashingType.TypeByteLength() * uint64(len(a))
}

func (a *ShardProposerSlashings) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li ShardProposerSlashings) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, spec.MAX_SHARD_PROPOSER_SLASHINGS)
}
