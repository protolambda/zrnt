package common

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"

	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Hash32 = Root

const Hash32Type = RootType

const MAX_EXTRA_DATA_BYTES = 32

var ExtraDataType = BasicListType(Uint8Type, MAX_EXTRA_DATA_BYTES)

type ExtraData []byte

func (otx *ExtraData) Deserialize(dr *codec.DecodingReader) error {
	return dr.ByteList((*[]byte)(otx), MAX_EXTRA_DATA_BYTES)
}

func (otx ExtraData) Serialize(w *codec.EncodingWriter) error {
	return w.Write(otx)
}

func (otx ExtraData) ByteLength() (out uint64) {
	return uint64(len(otx))
}

func (otx *ExtraData) FixedLength() uint64 {
	return 0
}

func (otx ExtraData) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.ByteListHTR(otx, MAX_EXTRA_DATA_BYTES)
}

func (otx ExtraData) MarshalText() ([]byte, error) {
	return conv.BytesMarshalText(otx[:])
}

func (otx ExtraData) String() string {
	return "0x" + hex.EncodeToString(otx[:])
}

func (otx *ExtraData) UnmarshalText(text []byte) error {
	if otx == nil {
		return errors.New("cannot decode into nil opaque transaction")
	}
	return conv.DynamicBytesUnmarshalText((*[]byte)(otx), text[:])
}

func (otx ExtraData) View() (*ExtraDataView, error) {
	dec := codec.NewDecodingReader(bytes.NewReader(otx), uint64(len(otx)))
	return AsExtraData(ExtraDataType.Deserialize(dec))
}

type ExtraDataView struct {
	*BasicListView
}

func AsExtraData(v View, err error) (*ExtraDataView, error) {
	c, err := AsBasicList(v, err)
	return &ExtraDataView{c}, err
}
func (v *ExtraDataView) Raw() (ExtraData, error) {
	var buf bytes.Buffer
	w := codec.NewEncodingWriter(&buf)
	if err := v.Serialize(w); err != nil {
		return nil, err
	}
	return ExtraData(buf.Bytes()), nil
}

type ExecutionEngine interface {
	ExecutePayload(ctx context.Context, executionPayload interface{}) (valid bool, err error)
	// TODO: remaining interface parts
}
