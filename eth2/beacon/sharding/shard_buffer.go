package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func ShardColumnType(spec *common.Spec) *ComplexListTypeDef {
	return ComplexListType(ShardWorkType(spec), spec.MAX_SHARDS)
}

type ShardColumnView struct{ *ComplexListView }

func AsShardColumn(v View, err error) (*ShardColumnView, error) {
	c, err := AsComplexList(v, err)
	return &ShardColumnView{c}, err
}

func (v *ShardColumnView) GetWork(shard common.Shard) (*ShardWorkView, error) {
	return AsShardWork(v.Get(uint64(shard)))
}

type ShardColumn []ShardWork

func (sc *ShardColumn) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*sc)
		*sc = append(*sc, ShardWork{})
		return spec.Wrap(&((*sc)[i]))
	}, 0, spec.MAX_SHARDS)
}

func (sc ShardColumn) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return spec.Wrap(&sc[i])
	}, 0, uint64(len(sc)))
}

func (sc ShardColumn) ByteLength(spec *common.Spec) (out uint64) {
	for _, v := range sc {
		out += v.ByteLength(spec) + codec.OFFSET_SIZE
	}
	return
}

func (sc *ShardColumn) FixedLength(*common.Spec) uint64 {
	return 0
}

func (sc ShardColumn) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(sc))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return spec.Wrap(&sc[i])
		}
		return nil
	}, length, spec.MAX_SHARDS)
}

func (sc ShardColumn) View(spec *common.Spec) (*ShardColumnView, error) {
	elements := make([]View, len(sc), len(sc))
	for i := 0; i < len(sc); i++ {
		v, err := sc[i].View(spec)
		if err != nil {
			return nil, err
		}
		elements[i] = v
	}
	return AsShardColumn(ShardColumnType(spec).FromElements(elements...))
}

func ShardBufferType(spec *common.Spec) *ComplexVectorTypeDef {
	return ComplexVectorType(ShardColumnType(spec), uint64(spec.SHARD_STATE_MEMORY_SLOTS))
}

type ShardBufferView struct{ *ComplexVectorView }

func AsShardBuffer(v View, err error) (*ShardBufferView, error) {
	c, err := AsComplexVector(v, err)
	return &ShardBufferView{c}, err
}

func (v *ShardBufferView) Column(i uint64) (*ShardColumnView, error) {
	return AsShardColumn(v.Get(i))
}

func (v *ShardBufferView) SetColumn(i uint64, view *ShardColumnView) error {
	return v.Set(i, view)
}

type ShardBuffer []ShardColumn

func (li *ShardBuffer) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	*li = make([]ShardColumn, spec.SHARD_STATE_MEMORY_SLOTS, spec.SHARD_STATE_MEMORY_SLOTS)
	return dr.Vector(func(i uint64) codec.Deserializable {
		return spec.Wrap(&(*li)[i])
	}, 0, uint64(spec.SHARD_STATE_MEMORY_SLOTS))
}

func (a ShardBuffer) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.Vector(func(i uint64) codec.Serializable {
		return spec.Wrap(&a[i])
	}, 0, uint64(spec.SHARD_STATE_MEMORY_SLOTS))
}

func (a ShardBuffer) ByteLength(spec *common.Spec) (out uint64) {
	out = uint64(len(a)) * codec.OFFSET_SIZE
	for _, v := range a {
		out += v.ByteLength(spec)
	}
	return
}

func (a *ShardBuffer) FixedLength(spec *common.Spec) uint64 {
	return 0
}

func (li ShardBuffer) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		return spec.Wrap(&li[i])
	}, uint64(spec.SHARD_STATE_MEMORY_SLOTS))
}
