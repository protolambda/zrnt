package hashing

import (
	"encoding/binary"
	"github.com/minio/sha256-simd"
	"github.com/protolambda/zssz/htr"
	"hash"
)

// Hash the given input. Use a new hashing object, and ditch it after hashing.
// See GetHashFn for a more efficient approach for repeated hashing.
// Defaults to SHA-256.
var Hash HashFn = sha256.Sum256

// Hashes the input, and returns the hash as a byte slice
type HashFn = htr.HashFn

type NewHashFn func() HashFn

// re-uses the sha256 working variables for each new call of a allocated hash-function.
func Sha256Repeat() HashFn {
	h := sha256.New()
	hashFn := func(in []byte) (out [32]byte) {
		h.Reset()
		h.Write(in)
		copy(out[:], h.Sum(nil))
		return
	}
	return hashFn
}

// Get a hash-function that re-uses the hashing working-variables. Defaults to SHA-256.
var GetHashFn NewHashFn = Sha256Repeat

func XorBytes32(a [32]byte, b [32]byte) (out [32]byte) {
	for i := 0; i < 32; i++ {
		out[i] = a[i] ^ b[i]
	}
	return
}

type MerkleFn = htr.MerkleFn

type NewMerkleFn func() MerkleFn

type sha256Scratch struct {
	inst    hash.Hash
	scratch [64]byte
}

func (s *sha256Scratch) Combi(a [32]byte, b [32]byte) (out [32]byte) {
	copy(s.scratch[:32], a[:])
	copy(s.scratch[32:], b[:])
	s.inst.Reset()
	s.inst.Write(s.scratch[:])
	copy(out[:], s.inst.Sum(nil))
	return
}

func (s *sha256Scratch) MixIn(a [32]byte, i uint64) (out [32]byte) {
	copy(s.scratch[:32], a[:])
	copy(s.scratch[32:], make([]byte, 32, 32))
	binary.LittleEndian.PutUint64(s.scratch[32:], i)
	s.inst.Reset()
	s.inst.Write(s.scratch[:])
	copy(out[:], s.inst.Sum(nil))
	return
}

func Sha256Merkle() MerkleFn {
	return &sha256Scratch{inst: sha256.New()}
}

// Get a merkle-function that re-uses the hashing working-variables as well as the merkle work variables.
// Defaults to SHA-256.
var GetMerkleFn NewMerkleFn = Sha256Merkle
