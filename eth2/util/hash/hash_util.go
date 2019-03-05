package hash

import (
	"crypto/sha256"
	"go-beacon-transition/eth2"
)

func Hash(input []byte) (out eth2.Bytes32) {
	// TODO this could be optimized,
	//  in reality you don't want to re-init the hashing function every time you call this
	hash := sha256.New()
	hash.Write(input)
	copy(out[:], hash.Sum(nil))
	return out
}

func XorBytes32(a eth2.Bytes32, b eth2.Bytes32) (out eth2.Bytes32) {
	for i := 0; i < 32; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}
