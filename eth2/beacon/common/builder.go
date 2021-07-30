package common

import (
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

const BuilderIndexType = Uint64Type

// Index of a builder, pointing to a builder registry location
type BuilderIndex Uint64View

func AsBuilderIndex(v View, err error) (BuilderIndex, error) {
	i, err := AsUint64(v, err)
	return BuilderIndex(i), err
}

func (bi *BuilderIndex) Deserialize(dr *codec.DecodingReader) error {
	return (*Uint64View)(bi).Deserialize(dr)
}

func (bi BuilderIndex) Serialize(w *codec.EncodingWriter) error {
	return w.WriteUint64(uint64(bi))
}

func (BuilderIndex) ByteLength() uint64 {
	return 8
}

func (BuilderIndex) FixedLength() uint64 {
	return 8
}

func (bi BuilderIndex) HashTreeRoot(hFn tree.HashFn) Root {
	return Uint64View(bi).HashTreeRoot(hFn)
}

func (bi BuilderIndex) MarshalJSON() ([]byte, error) {
	return Uint64View(bi).MarshalJSON()
}

func (bi *BuilderIndex) UnmarshalJSON(b []byte) error {
	return ((*Uint64View)(bi)).UnmarshalJSON(b)
}

func (bi BuilderIndex) String() string {
	return Uint64View(bi).String()
}
