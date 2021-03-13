package altair

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type LeakScores []Uint64View

func (a *LeakScores) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, Uint64View(0))
		return &(*a)[i]
	}, Uint64Type.TypeByteLength(), spec.VALIDATOR_REGISTRY_LIMIT)
}

func (a LeakScores) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, Uint64Type.TypeByteLength(), uint64(len(a)))
}

func (a LeakScores) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(a)) * Uint64Type.TypeByteLength()
}

func (a *LeakScores) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li LeakScores) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(li[i])
	}, length, spec.VALIDATOR_REGISTRY_LIMIT)
}

func LeakScoresType(spec *common.Spec) *BasicListTypeDef {
	return BasicListType(Uint64Type, spec.VALIDATOR_REGISTRY_LIMIT)
}

type LeakScoresView struct {
	*BasicListView
}

func AsLeakScores(v View, err error) (*LeakScoresView, error) {
	c, err := AsBasicList(v, err)
	return &LeakScoresView{c}, err
}

func (v *LeakScoresView) GetScore(index common.ValidatorIndex) (Uint64View, error) {
	return v.Get(uint64(index))
}

func (v *LeakScoresView) SetScore(index common.ValidatorIndex, score Uint64View) error {
	return v.Set(uint64(index), score)
}
