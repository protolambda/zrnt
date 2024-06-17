package common

import (
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Number Uint64View

const NumberType = Uint64Type

func (a *Number) Deserialize(dr *codec.DecodingReader) error {
	return (*Uint64View)(a).Deserialize(dr)
}

func (i Number) Serialize(w *codec.EncodingWriter) error {
	return w.WriteUint64(uint64(i))
}

func (Number) ByteLength() uint64 {
	return 8
}

func (Number) FixedLength() uint64 {
	return 8
}

func (t Number) HashTreeRoot(hFn tree.HashFn) Root {
	return Uint64View(t).HashTreeRoot(hFn)
}

func (e Number) MarshalJSON() ([]byte, error) {
	return Uint64View(e).MarshalJSON()
}

func (e *Number) UnmarshalJSON(b []byte) error {
	return ((*Uint64View)(e)).UnmarshalJSON(b)
}

func (e Number) String() string {
	return Uint64View(e).String()
}

func AsNumber(v View, err error) (Number, error) {
	i, err := AsUint64(v, err)
	return Number(i), err
}
