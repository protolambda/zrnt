package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func PendingShardHeaderType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("PendingShardHeader", []FieldDef{
		{"commitment", DataCommitmentType},
		{"root", RootType},
		{"votes", phase0.AttestationBitsType(spec)},
		{"weight", common.GweiType},
	})
}

type PendingShardHeader struct {
	// KZG10 commitment to the data
	Commitment DataCommitment `json:"commitment" yaml:"commitment"`
	// hash_tree_root of the ShardHeader (stored so that attestations can be checked against it)
	Root common.Root `json:"root" yaml:"root"`
	// Who voted for the header
	Votes phase0.AttestationBits `json:"votes" yaml:"votes"`
	// Sum of effective balances of votes
	Weight common.Gwei `json:"weight" yaml:"weight"`
}

func (h *PendingShardHeader) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.Container(&h.Commitment, &h.Root, spec.Wrap(&h.Votes), &h.Weight)
}

func (h *PendingShardHeader) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Container(&h.Commitment, &h.Root, spec.Wrap(&h.Votes), &h.Weight)
}

func (h *PendingShardHeader) ByteLength(spec *common.Spec) uint64 {
	return DataCommitmentType.TypeByteLength() + 32 + h.Votes.ByteLength(spec) + 8
}

func (h *PendingShardHeader) FixedLength(spec *common.Spec) uint64 {
	return DataCommitmentType.TypeByteLength() + 32 + 4 + 8
}

func (h *PendingShardHeader) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&h.Commitment, &h.Root, spec.Wrap(&h.Votes), &h.Weight)
}

func PendingShardHeadersType(spec *common.Spec) *ComplexListTypeDef {
	return ComplexListType(PendingShardHeaderType(spec), spec.MAX_SHARD_HEADERS_PER_SHARD)
}

type PendingShardHeaders []PendingShardHeader

func (hl *PendingShardHeaders) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*hl)
		*hl = append(*hl, PendingShardHeader{})
		return spec.Wrap(&((*hl)[i]))
	}, 0, spec.MAX_SHARD_HEADERS_PER_SHARD)
}

func (hl PendingShardHeaders) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&hl[i])
	}, 0, uint64(len(hl)))
}

func (hl PendingShardHeaders) ByteLength(spec *common.Spec) (out uint64) {
	for _, v := range hl {
		out += v.ByteLength(spec) + codec.OFFSET_SIZE
	}
	return
}

func (hl *PendingShardHeaders) FixedLength(*common.Spec) uint64 {
	return 0
}

func (hl PendingShardHeaders) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(hl))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&hl[i])
		}
		return nil
	}, length, spec.MAX_SHARD_HEADERS_PER_SHARD)
}
