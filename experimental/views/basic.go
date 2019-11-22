package views

import (
	"encoding/binary"
	. "github.com/protolambda/zrnt/experimental/tree"
)

// A basic type, identified by its size in bits.
type BasicType uint64

func (cd BasicType) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (cd BasicType) ByteLength() uint64 {
	length := uint64(cd) >> 8
	if length == 0 {
		return 1
	}
	return length
}

func (cd BasicType) ViewFromBacking(node Node) View {
	v, ok := node.(*Root)
	if !ok {
		return nil
	}
	switch cd {
	case 1:
		return BoolView(*v != (Root{}))
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

func (cd BasicType) SubViewFromBacking(v *Root, i uint8) SubView {
	if uint64(i) >= (256 / uint64(cd)) {
		return nil
	}
	switch cd {
	case 1:
		return BoolView(v[i>>3]>>(i&7) == 1)
	case 8:
		return Uint8View(v[i])
	case 16:
		return Uint16View(binary.LittleEndian.Uint16(v[2*i:2*i+2]))
	case 32:
		return Uint32View(binary.LittleEndian.Uint32(v[4*i:4*i+4]))
	case 64:
		return Uint64View(binary.LittleEndian.Uint64(v[8*i:8*i+8]))
	default:
		// unsupported backing
		return nil
	}
}

const (
	BoolType   BasicType = 1
	Uint8Type  BasicType = 8
	Uint16Type BasicType = 16
	Uint32Type BasicType = 32
	Uint64Type BasicType = 64
)

type Uint8View uint8

func (v Uint8View) Backing() Node {
	out := &Root{}
	out[0] = uint8(v)
	return out
}

func (v Uint8View) BackingFromBase(base *Root, i uint8) *Root {
	if i >= 32 {
		return nil
	}
	newRoot := *base
	newRoot[i] = uint8(v)
	return &newRoot
}

type Uint16View uint16

func (v Uint16View) Backing() Node {
	out := &Root{}
	binary.LittleEndian.PutUint16(out[:2], uint16(v))
	return out
}

func (v Uint16View) BackingFromBase(base *Root, i uint8) *Root {
	if i >= 16 {
		return nil
	}
	newRoot := *base
	binary.LittleEndian.PutUint16(newRoot[i*2:i*2+2], uint16(v))
	return &newRoot
}

type Uint32View uint32

func (v Uint32View) Backing() Node {
	out := &Root{}
	binary.LittleEndian.PutUint32(out[:4], uint32(v))
	return out
}

func (v Uint32View) BackingFromBase(base *Root, i uint8) *Root {
	if i >= 8 {
		return nil
	}
	newRoot := *base
	binary.LittleEndian.PutUint32(newRoot[i*4:i*4+4], uint32(v))
	return &newRoot
}

type Uint64View uint64

func (v Uint64View) Backing() Node {
	out := &Root{}
	binary.LittleEndian.PutUint64(out[:8], uint64(v))
	return out
}

func (v Uint64View) BackingFromBase(base *Root, i uint8) *Root {
	if i >= 4 {
		return nil
	}
	newRoot := *base
	binary.LittleEndian.PutUint64(newRoot[i*8:i*8+8], uint64(v))
	return &newRoot
}

type BoolView bool

var trueRoot = &Root{1}

func (v BoolView) Backing() Node {
	if v {
		return trueRoot
	} else {
		return &ZeroHashes[0]
	}
}

func (v BoolView) BackingFromBase(base *Root, i uint8) *Root {
	// copy the old root, do not mutate the immutable.
	newRoot := *base
	if v {
		newRoot[i>>3] |= 1 << (i & 7)
	} else {
		newRoot[i>>3] &^= 1 << (i & 7)
	}
	return &newRoot
}
