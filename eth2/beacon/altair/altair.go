package altair

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type ParticipationFlags Uint8View

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

// Participation flag fractions
const (
	TIMELY_HEAD_FLAG_NUMERATOR   uint64 = 12
	TIMELY_SOURCE_FLAG_NUMERATOR uint64 = 12
	TIMELY_TARGET_FLAG_NUMERATOR uint64 = 32
	FLAG_DENOMINATOR             uint64 = 64
)

// TODO Altair block types
// TODO Altair state type
