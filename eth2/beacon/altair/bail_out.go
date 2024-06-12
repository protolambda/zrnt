package altair

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func BlockBailOutsType(spec *common.Spec) ListTypeDef {
	return ListType(BailOutType, uint64(spec.MAX_VOLUNTARY_EXITS))
}

type BailOuts []BailOut

func (a *BailOuts) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, BailOut{})
		return &(*a)[i]
	}, BailOutType.TypeByteLength(), uint64(spec.MAX_VOLUNTARY_EXITS))
}

func (a BailOuts) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, BailOutType.TypeByteLength(), uint64(len(a)))
}

func (a BailOuts) ByteLength(spec *common.Spec) (out uint64) {
	return BailOutType.TypeByteLength() * uint64(len(a))
}

func (*BailOuts) FixedLength(*common.Spec) uint64 {
	return 0
}

func (li BailOuts) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, uint64(spec.MAX_VOLUNTARY_EXITS))
}

type BailOut struct {
	// Earliest epoch when voluntary exit can be processed
	ValidatorIndex common.ValidatorIndex `json:"validator_index" yaml:"validator_index"`
}

var BailOutType = ContainerType("BailOut", []FieldDef{
	{"validator_index", common.ValidatorIndexType},
})

func (v *BailOut) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&v.ValidatorIndex)
}

func (v *BailOut) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&v.ValidatorIndex)
}

func (v *BailOut) ByteLength() uint64 {
	return BailOutType.TypeByteLength()
}

func (*BailOut) FixedLength() uint64 {
	return BailOutType.TypeByteLength()
}

func (v *BailOut) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(v.ValidatorIndex)
}
