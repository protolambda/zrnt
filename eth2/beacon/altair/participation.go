package altair

import (
	"context"
	"encoding/binary"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

const ParticipationFlagsType = Uint8Type

type ParticipationFlags Uint8View

func AsParticipationFlags(v View, err error) (ParticipationFlags, error) {
	i, err := AsUint8(v, err)
	return ParticipationFlags(i), err
}

func (a *ParticipationFlags) Deserialize(dr *codec.DecodingReader) error {
	return (*Uint8View)(a).Deserialize(dr)
}

func (i ParticipationFlags) Serialize(w *codec.EncodingWriter) error {
	return w.WriteByte(uint8(i))
}

func (ParticipationFlags) ByteLength() uint64 {
	return 8
}

func (ParticipationFlags) FixedLength() uint64 {
	return 8
}

func (t ParticipationFlags) HashTreeRoot(hFn tree.HashFn) common.Root {
	return Uint8View(t).HashTreeRoot(hFn)
}

func (e ParticipationFlags) String() string {
	return Uint8View(e).String()
}

// Participation flag indices
const (
	TIMELY_HEAD_FLAG_INDEX   uint8 = 0
	TIMELY_SOURCE_FLAG_INDEX uint8 = 1
	TIMELY_TARGET_FLAG_INDEX uint8 = 2
)

const (
	TIMELY_HEAD_FLAG   ParticipationFlags = 1 << TIMELY_HEAD_FLAG_INDEX
	TIMELY_SOURCE_FLAG ParticipationFlags = 1 << TIMELY_SOURCE_FLAG_INDEX
	TIMELY_TARGET_FLAG ParticipationFlags = 1 << TIMELY_TARGET_FLAG_INDEX
)

// Participation flag fractions
const (
	TIMELY_HEAD_WEIGHT   uint64 = 12
	TIMELY_SOURCE_WEIGHT uint64 = 12
	TIMELY_TARGET_WEIGHT uint64 = 32
	SYNC_REWARD_WEIGHT   uint64 = 8
	WEIGHT_DENOMINATOR   uint64 = 64
)

type ParticipationRegistry []ParticipationFlags

func (a *ParticipationRegistry) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(*a)
		*a = append(*a, ParticipationFlags(0))
		return &(*a)[i]
	}, ParticipationFlagsType.TypeByteLength(), spec.VALIDATOR_REGISTRY_LIMIT)
}

func (a ParticipationRegistry) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.List(func(i uint64) codec.Serializable {
		return &a[i]
	}, ParticipationFlagsType.TypeByteLength(), uint64(len(a)))
}

func (a ParticipationRegistry) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(a)) * ParticipationFlagsType.TypeByteLength()
}

func (a *ParticipationRegistry) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li ParticipationRegistry) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.Uint8ListHTR(func(i uint64) uint8 {
		return uint8(li[i])
	}, uint64(len(li)), spec.VALIDATOR_REGISTRY_LIMIT)
}

func ParticipationRegistryType(spec *common.Spec) *BasicListTypeDef {
	return BasicListType(ParticipationFlagsType, spec.VALIDATOR_REGISTRY_LIMIT)
}

type ParticipationRegistryView struct {
	*BasicListView
}

func AsParticipationRegistry(v View, err error) (*ParticipationRegistryView, error) {
	c, err := AsBasicList(v, err)
	return &ParticipationRegistryView{c}, err
}

func (v *ParticipationRegistryView) GetFlags(index common.ValidatorIndex) (ParticipationFlags, error) {
	return AsParticipationFlags(v.Get(uint64(index)))
}

func (v *ParticipationRegistryView) SetFlags(index common.ValidatorIndex, score ParticipationFlags) error {
	return v.Set(uint64(index), Uint8View(score))
}

func (v *ParticipationRegistryView) FillZeroes(length uint64) error {
	// 32 flags (uint8) per node (bytes32)
	nodesLen := (length + 31) / 32
	depth := tree.CoverDepth(v.BottomNodeLimit())
	zero := &tree.Root{}
	contents, err := tree.SubtreeFillToLength(zero, depth, nodesLen)
	if err != nil {
		return err
	}
	lengthNode := &tree.Root{}
	binary.LittleEndian.PutUint64(lengthNode[:8], length)
	return v.SetBacking(tree.NewPairNode(contents, lengthNode))
}

func ProcessParticipationFlagUpdates(ctx context.Context, spec *common.Spec, state *BeaconStateView) error {
	currentEp, err := state.CurrentEpochParticipation()
	if err != nil {
		return err
	}
	previousEp, err := state.PreviousEpochParticipation()
	if err != nil {
		return err
	}
	if err := previousEp.SetBacking(currentEp.Backing()); err != nil {
		return err
	}
	length, err := currentEp.Length()
	if err != nil {
		return err
	}
	return currentEp.FillZeroes(length)
}
