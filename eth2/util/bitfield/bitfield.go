package bitfield

// bits are indexed from left to right of internal byte array (like a little endian integer).
// But inside byte it is from right to left. //TODO: correct?
type Bitfield []byte

func (b Bitfield) GetBit(i uint64) byte {
	if uint64(len(b))<<3 > i {
		return (b[i>>3] >> (i & 7)) & 1
	}
	panic("invalid bitfield access")
}

// Verify bitfield against the size:
//  - the bitfield must have the correct amount of bytes
//  - bits after this size (in bits) must be 0.
func (b Bitfield) VerifySize(size uint64) bool {
	// check byte count
	if uint64(len(b)) != (size+7)>>3 {
		return false
	}
	// check if bitfield is padded with zero bits only
	end := uint64(len(b)) << 3
	for i := size; i < end; i++ {
		if b.GetBit(i) == 1 {
			return false
		}
	}
	return true
}

func (b Bitfield) IsZero() bool {
	for i := 0; i < len(b); i++ {
		if b[i] != 0 {
			return false
		}
	}
	return true
}
