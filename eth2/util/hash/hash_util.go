package hash

import (
	"crypto/sha256"
)

func Hash(input []byte) (out [32]byte) {
	// TODO this could be optimized,
	//  in reality you don't want to re-init the hashing function every time you call this
	hash := sha256.New()
	hash.Write(input)
	copy(out[:], hash.Sum(nil))
	return out
}

func XorBytes32(a [32]byte, b [32]byte) (out [32]byte) {
	for i := 0; i < 32; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}
