package altair

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type InactivityScores []Uint64View

func (a *InactivityScores) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, Uint64View(0))
		return &(*a)[i]
	}, Uint64Type.TypeByteLength(), spec.VALIDATOR_REGISTRY_LIMIT)
}

func (a InactivityScores) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, Uint64Type.TypeByteLength(), uint64(len(a)))
}

func (a InactivityScores) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(a)) * Uint64Type.TypeByteLength()
}

func (a *InactivityScores) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li InactivityScores) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(li[i])
	}, length, spec.VALIDATOR_REGISTRY_LIMIT)
}

func InactivityScoresType(spec *common.Spec) *BasicListTypeDef {
	return BasicListType(Uint64Type, spec.VALIDATOR_REGISTRY_LIMIT)
}

type InactivityScoresView struct {
	*BasicListView
}

func AsInactivityScores(v View, err error) (*InactivityScoresView, error) {
	c, err := AsBasicList(v, err)
	return &InactivityScoresView{c}, err
}

func (v *InactivityScoresView) GetScore(index common.ValidatorIndex) (Uint64View, error) {
	return AsUint64(v.Get(uint64(index)))
}

func (v *InactivityScoresView) SetScore(index common.ValidatorIndex, score Uint64View) error {
	return v.Set(uint64(index), score)
}
