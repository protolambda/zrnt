package common

import (
	"encoding/hex"
	"errors"

	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

const Eth1AddressType = SmallByteVecMeta(20)

func AsEth1Address(v View, err error) (Eth1Address, error) {
	c, err := AsSmallByteVec(v, err)
	if err != nil {
		return Eth1Address{}, err
	}
	var out Eth1Address
	copy(out[:], c)
	return out, nil
}

type Eth1Address [20]byte

func (addr Eth1Address) View() SmallByteVecView {
	// overlay on a copy of the value
	return SmallByteVecView(addr[:])
}

func (addr *Eth1Address) Deserialize(dr *codec.DecodingReader) error {
	if addr == nil {
		return errors.New("cannot deserialize into nil eth1 address")
	}
	_, err := dr.Read(addr[:])
	return err
}

func (addr *Eth1Address) Serialize(w *codec.EncodingWriter) error {
	return w.Write(addr[:])
}

func (*Eth1Address) ByteLength() uint64 {
	return 20
}

func (*Eth1Address) FixedLength() uint64 {
	return 20
}

func (addr *Eth1Address) HashTreeRoot(hFn tree.HashFn) tree.Root {
	var out tree.Root
	copy(out[0:20], addr[:])
	return out
}

func (addr Eth1Address) MarshalText() ([]byte, error) {
	return conv.BytesMarshalText(addr[:])
}

func (addr Eth1Address) String() string {
	return "0x" + hex.EncodeToString(addr[:])
}

func (addr *Eth1Address) UnmarshalText(text []byte) error {
	if addr == nil {
		return errors.New("cannot decode into nil Eth1Address")
	}
	return conv.FixedBytesUnmarshalText(addr[:], text[:])
}
