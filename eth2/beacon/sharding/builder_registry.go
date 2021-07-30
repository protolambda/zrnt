package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type RegistryIndices []common.BuilderIndex

func (p *RegistryIndices) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*p)
		*p = append(*p, common.BuilderIndex(0))
		return &((*p)[i])
	}, common.BuilderIndexType.TypeByteLength(), spec.VALIDATOR_REGISTRY_LIMIT)
}

func (a RegistryIndices) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return a[i]
	}, common.BuilderIndexType.TypeByteLength(), uint64(len(a)))
}

func (a RegistryIndices) ByteLength(spec *common.Spec) uint64 {
	return common.BuilderIndexType.TypeByteLength() * uint64(len(a))
}

func (*RegistryIndices) FixedLength() uint64 {
	return 0
}

func (p RegistryIndices) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.Uint64ListHTR(func(i uint64) uint64 {
		return uint64(p[i])
	}, uint64(len(p)), spec.VALIDATOR_REGISTRY_LIMIT)
}

type BuilderRegistry []*Builder

func (a *BuilderRegistry) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, &Builder{})
		return (*a)[i]
	}, BuilderType.TypeByteLength(), spec.VALIDATOR_REGISTRY_LIMIT)
}

func (a BuilderRegistry) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return a[i]
	}, BuilderType.TypeByteLength(), uint64(len(a)))
}

func (a BuilderRegistry) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(a)) * BuilderType.TypeByteLength()
}

func (a *BuilderRegistry) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li BuilderRegistry) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return li[i]
		}
		return nil
	}, length, spec.VALIDATOR_REGISTRY_LIMIT)
}

func BuildersRegistryType(spec *common.Spec) ListTypeDef {
	return ComplexListType(BuilderType, spec.VALIDATOR_REGISTRY_LIMIT)
}

type BuildersRegistryView struct{ *ComplexListView }

func AsBuildersRegistry(v View, err error) (*BuildersRegistryView, error) {
	c, err := AsComplexList(v, err)
	return &BuildersRegistryView{c}, nil
}

func (registry *BuildersRegistryView) BuilderCount() (uint64, error) {
	return registry.Length()
}

func (registry *BuildersRegistryView) Builder(index common.BuilderIndex) (common.Builder, error) {
	return AsBuilder(registry.Get(uint64(index)))
}

func (registry *BuildersRegistryView) Iter() (next func() (val common.Builder, ok bool, err error)) {
	iter := registry.ReadonlyIter()
	return func() (val common.Builder, ok bool, err error) {
		elem, ok, err := iter.Next()
		if err != nil || !ok {
			return nil, ok, err
		}
		v, err := AsBuilder(elem, nil)
		return v, true, err
	}
}

func (registry *BuildersRegistryView) IsValidIndex(index common.BuilderIndex) (valid bool, err error) {
	count, err := registry.Length()
	if err != nil {
		return false, err
	}
	return uint64(index) < count, nil
}
