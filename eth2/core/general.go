package core

type Root [32]byte

type Bytes []byte

// 32 bits, not strictly an integer, hence represented as 4 bytes
// (bytes not necessarily corresponding to versions)
type ForkVersion [4]byte

func (v ForkVersion) ToUint32() uint32 {
	return uint32(v[0])<<24 | uint32(v[1])<<16 | uint32(v[2])<<8 | uint32(v[3])
}

type Shard uint64

type Gwei uint64
