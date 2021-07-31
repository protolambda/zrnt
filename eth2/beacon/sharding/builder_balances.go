package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type BuilderBalances []common.Gwei

func (a *BuilderBalances) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, common.Gwei(0))
		return &(*a)[i]
	}, common.GweiType.TypeByteLength(), spec.BLOB_BUILDER_REGISTRY_LIMIT)
}

func (a BuilderBalances) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, common.GweiType.TypeByteLength(), uint64(len(a)))
}

func (a BuilderBalances) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(a)) * common.GweiType.TypeByteLength()
}

func (a *BuilderBalances) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li BuilderBalances) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(li[i])
	}, length, spec.BLOB_BUILDER_REGISTRY_LIMIT)
}

func (li BuilderBalances) View(limit uint64) (*BuilderRegistryBalancesView, error) {
	// TODO: bad copy, converting to a tree more directly somehow would be nice.
	tmp := make([]BasicView, len(li), len(li))
	for i, bal := range li {
		tmp[i] = Uint64View(bal)
	}
	typ := BasicListType(common.GweiType, limit)
	return AsBuilderRegistryBalances(typ.FromElements(tmp...))
}

func BuilderRegistryBalancesType(spec *common.Spec) *BasicListTypeDef {
	return BasicListType(common.GweiType, spec.BLOB_BUILDER_REGISTRY_LIMIT)
}

type BuilderRegistryBalancesView struct {
	*BasicListView
}

var _ common.BuilderBalancesRegistry = (*BuilderRegistryBalancesView)(nil)

func AsBuilderRegistryBalances(v View, err error) (*BuilderRegistryBalancesView, error) {
	c, err := AsBasicList(v, err)
	return &BuilderRegistryBalancesView{c}, err
}

func (v *BuilderRegistryBalancesView) GetBalance(index common.BuilderIndex) (common.Gwei, error) {
	return common.AsGwei(v.Get(uint64(index)))
}

func (v *BuilderRegistryBalancesView) SetBalance(index common.BuilderIndex, bal common.Gwei) error {
	return v.Set(uint64(index), Uint64View(bal))
}

func (v *BuilderRegistryBalancesView) AppendBalance(bal common.Gwei) error {
	return v.Append(Uint64View(bal))
}

func (v *BuilderRegistryBalancesView) AllBalances() ([]common.Gwei, error) {
	var out []common.Gwei
	balIter := v.ReadonlyIter()
	for i := common.BuilderIndex(0); true; i++ {
		el, ok, err := balIter.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		balance, err := common.AsGwei(el, nil)
		if err != nil {
			return nil, err
		}
		out = append(out, balance)
	}
	return out, nil
}

func (v *BuilderRegistryBalancesView) Iter() (next func() (bal common.Gwei, ok bool, err error)) {
	iter := v.ReadonlyIter()
	return func() (bal common.Gwei, ok bool, err error) {
		el, ok, err := iter.Next()
		if err != nil || !ok {
			return 0, ok, err
		}
		v, err := common.AsGwei(el, err)
		return v, true, err
	}
}
