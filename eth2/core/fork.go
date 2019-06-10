package core

// 32 bits, not strictly an integer, hence represented as 4 bytes
// (bytes not necessarily corresponding to versions)
type ForkVersion [4]byte

func (v ForkVersion) ToUint32() uint32 {
	return uint32(v[0])<<24 | uint32(v[1])<<16 | uint32(v[2])<<8 | uint32(v[3])
}

func Int32ToForkVersion(v uint32) ForkVersion {
	return [4]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
}
