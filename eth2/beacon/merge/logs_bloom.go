package merge

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	"github.com/protolambda/ztyp/tree"
	"github.com/protolambda/ztyp/view"
)

const BYTES_PER_LOGS_BLOOM = 256

func AsLogsBloom(v view.View, err error) (*LogsBloom, error) {
	var out LogsBloom
	buf := codec.NewEncodingWriter(bytes.NewBuffer(out[:]))
	if err := v.Serialize(buf); err != nil {
		return &out, nil
	}
	if x := buf.Written(); x != BYTES_PER_LOGS_BLOOM {
		return nil, fmt.Errorf("unexpected logs bloom tree view, got %d bytes", x)
	}
	return &out, nil
}

var LogsBloomType = view.BasicVectorType(view.Uint8Type, BYTES_PER_LOGS_BLOOM)

type LogsBloom [BYTES_PER_LOGS_BLOOM]byte

func (s *LogsBloom) Deserialize(dr *codec.DecodingReader) error {
	if s == nil {
		return errors.New("cannot deserialize into nil logs bloom")
	}
	_, err := dr.Read(s[:])
	return err
}

func (s *LogsBloom) Serialize(w *codec.EncodingWriter) error {
	return w.Write(s[:])
}

func (*LogsBloom) ByteLength() uint64 {
	return BYTES_PER_LOGS_BLOOM
}

func (*LogsBloom) FixedLength() uint64 {
	return BYTES_PER_LOGS_BLOOM
}

func (s *LogsBloom) HashTreeRoot(hFn tree.HashFn) tree.Root {
	var bottom [8]tree.Root
	for i := 0; i < 8; i++ {
		copy(bottom[i][:], s[i<<5:(i+1)<<5])
	}
	a := hFn(bottom[0], bottom[1])
	b := hFn(bottom[2], bottom[3])
	c := hFn(bottom[4], bottom[5])
	d := hFn(bottom[6], bottom[7])
	return hFn(hFn(a, b), hFn(c, d))
}

func (p *LogsBloom) MarshalText() ([]byte, error) {
	return conv.BytesMarshalText(p[:])
}

func (p LogsBloom) String() string {
	return "0x" + hex.EncodeToString(p[:])
}

func (p *LogsBloom) UnmarshalText(text []byte) error {
	if p == nil {
		return errors.New("cannot decode into nil logs bloom")
	}
	return conv.FixedBytesUnmarshalText(p[:], text[:])
}
