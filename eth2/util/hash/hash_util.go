package hash

import (
	"crypto/sha256"
	. "github.com/protolambda/zrnt/eth2/core"
)

// hashes, and returns the hash as a root
func HashRoot(input []byte) (out Root) {
	copy(out[:], Hash(input))
	return
}

// Hash the given input. Use a new hashing object, and ditch it after hashing.
// See GetHashFn for a more efficient approach for repeated hashing.
func Hash(input []byte) (out []byte) {
	hash := sha256.New()
	hash.Write(input)
	return hash.Sum(nil)
}

// Hashes the input, and returns the hash as a byte slice
type HashFn func(input []byte) []byte

// Get a hash-function that re-uses the hashing working-variables
func GetHashFn() HashFn {
	hash := sha256.New()
	hashFn := func(in []byte) []byte {
		hash.Reset()
		hash.Write(in)
		return hash.Sum(nil)
	}
	return hashFn
}

func XorRoots(a Root, b Root) (out Root) {
	for i := 0; i < 32; i++ {
		out[i] = a[i] ^ b[i]
	}
	return
}
