package altair

import (
	"bytes"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type BailoutScores []Uint64View

func (a *BailoutScores) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, Uint64View(0))
		return &(*a)[i]
	}, Uint64Type.TypeByteLength(), uint64(spec.VALIDATOR_REGISTRY_LIMIT))
}

func (a BailoutScores) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, Uint64Type.TypeByteLength(), uint64(len(a)))
}

func (a BailoutScores) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(a)) * Uint64Type.TypeByteLength()
}

func (a *BailoutScores) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li BailoutScores) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(li[i])
	}, length, uint64(spec.VALIDATOR_REGISTRY_LIMIT))
}

func (li BailoutScores) View(spec *common.Spec) (*ParticipationRegistryView, error) {
	typ := BailoutScoresType(spec)
	var buf bytes.Buffer
	if err := li.Serialize(spec, codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	data := buf.Bytes()
	dec := codec.NewDecodingReader(bytes.NewReader(data), uint64(len(data)))
	return AsParticipationRegistry(typ.Deserialize(dec))
}

func BailoutScoresType(spec *common.Spec) *BasicListTypeDef {
	return BasicListType(Uint64Type, uint64(spec.VALIDATOR_REGISTRY_LIMIT))
}

type BailoutScoresView struct {
	*BasicListView
}

func AsBailoutScores(v View, err error) (*BailoutScoresView, error) {
	c, err := AsBasicList(v, err)
	return &BailoutScoresView{c}, err
}

func (v *BailoutScoresView) GetScore(index common.ValidatorIndex) (uint64, error) {
	s, err := AsUint64(v.Get(uint64(index)))
	return uint64(s), err
}

func (v *BailoutScoresView) SetScore(index common.ValidatorIndex, score uint64) error {
	return v.Set(uint64(index), Uint64View(score))
}
