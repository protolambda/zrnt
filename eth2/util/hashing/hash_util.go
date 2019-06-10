package hashing

import (
	"crypto/sha256"
)

// Hash the given input. Use a new hashing object, and ditch it after hashing.
// See GetHashFn for a more efficient approach for repeated hashing.
func Hash(input []byte) (out [32]byte) {
	return sha256.Sum256(input)
}

// Hashes the input, and returns the hash as a byte slice
type HashFn func(input []byte) [32]byte

// Get a hash-function that re-uses the hashing working-variables
func GetHashFn() HashFn {
	hash := sha256.New()
	hashFn := func(in []byte) (out [32]byte) {
		hash.Reset()
		hash.Write(in)
		copy(out[:], hash.Sum(nil))
		return
	}
	return hashFn
}

func XorBytes32(a [32]byte, b [32]byte) (out [32]byte) {
	for i := 0; i < 32; i++ {
		out[i] = a[i] ^ b[i]
	}
	return
}
