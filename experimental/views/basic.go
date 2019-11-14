package views

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type UintType uint64

func (cd UintType) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (cd UintType) ViewFromBacking(node Node) View {
	v := node.(*Root)
	switch cd {
	case 8:
		return Uint8View(v[0])
	case 16:
		return Uint16View(binary.LittleEndian.Uint16(v[:2]))
	case 32:
		return Uint32View(binary.LittleEndian.Uint32(v[:4]))
	case 64:
		return Uint64View(binary.LittleEndian.Uint64(v[:8]))
	default:
		// unsupported backing
		return nil
	}
}

const (
	Uint8Type  UintType = 8
	Uint16Type UintType = 16
	Uint32Type UintType = 32
	Uint64Type UintType = 64
)

type Uint8View uint8

func (v Uint8View) Backing() Node {
	out := &Root{}
	out[0] = uint8(v)
	return out
}

type Uint16View uint16

func (v Uint16View) Backing() Node {
	out := &Root{}
	binary.LittleEndian.PutUint16(out[:2], uint16(v))
	return out
}

type Uint32View uint32

func (v Uint32View) Backing() Node {
	out := &Root{}
	binary.LittleEndian.PutUint32(out[:4], uint32(v))
	return out
}

type Uint64View uint64

func (v Uint64View) Backing() Node {
	out := &Root{}
	binary.LittleEndian.PutUint64(out[:8], uint64(v))
	return out
}
