package merge

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type RegistryIndices []common.ValidatorIndex

func (p *RegistryIndices) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*p)
		*p = append(*p, common.ValidatorIndex(0))
		return &((*p)[i])
	}, common.ValidatorIndexType.TypeByteLength(), spec.WITHDRAWAL_REGISTRY_LIMIT)
}

func (a RegistryIndices) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return a[i]
	}, common.ValidatorIndexType.TypeByteLength(), uint64(len(a)))
}

func (a RegistryIndices) ByteLength(spec *common.Spec) uint64 {
	return common.ValidatorIndexType.TypeByteLength() * uint64(len(a))
}

func (*RegistryIndices) FixedLength() uint64 {
	return 0
}

func (p RegistryIndices) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(p[i])
	}, uint64(len(p)), spec.WITHDRAWAL_REGISTRY_LIMIT)
}

type WithdrawalRegistry []*Withdrawal

func (a *WithdrawalRegistry) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, &Withdrawal{})
		return (*a)[i]
	}, WithdrawalType.TypeByteLength(), spec.WITHDRAWAL_REGISTRY_LIMIT)
}

func (a WithdrawalRegistry) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return a[i]
	}, WithdrawalType.TypeByteLength(), uint64(len(a)))
}

func (a WithdrawalRegistry) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(a)) * WithdrawalType.TypeByteLength()
}

func (a *WithdrawalRegistry) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li WithdrawalRegistry) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return li[i]
		}
		return nil
	}, length, spec.WITHDRAWAL_REGISTRY_LIMIT)
}

func WithdrawalsRegistryType(spec *common.Spec) ListTypeDef {
	return ComplexListType(WithdrawalType, spec.WITHDRAWAL_REGISTRY_LIMIT)
}

type WithdrawalsRegistryView struct{ *ComplexListView }

func AsWithdrawalsRegistry(v View, err error) (*WithdrawalsRegistryView, error) {
	c, err := AsComplexList(v, err)
	return &WithdrawalsRegistryView{c}, nil
}

func (registry *WithdrawalsRegistryView) WithdrawalCount() (uint64, error) {
	return registry.Length()
}

func (registry *WithdrawalsRegistryView) Withdrawal(index common.ValidatorIndex) (*Withdrawal, error) {
	return AsWithdrawal(registry.Get(uint64(index)))
}

func (registry *WithdrawalsRegistryView) Iter() (next func() (val *Withdrawal, ok bool, err error)) {
	iter := registry.ReadonlyIter()
	return func() (val *Withdrawal, ok bool, err error) {
		elem, ok, err := iter.Next()
		if err != nil || !ok {
			return nil, ok, err
		}
		w, err := AsWithdrawal(elem, nil)
		return w, true, err
	}
}

func (registry *WithdrawalsRegistryView) IsValidIndex(index common.ValidatorIndex) (valid bool, err error) {
	count, err := registry.Length()
	if err != nil {
		return false, err
	}
	return uint64(index) < count, nil
}
