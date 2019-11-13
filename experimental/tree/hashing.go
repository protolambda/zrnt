package tree

import "github.com/minio/sha256-simd"

type Root [32]byte

type HashFn func(a Root, b Root) Root

func (h *HashFn) MerkleizeBytes(b []byte) Root {
	// TODO
	return Root{}
}

func (s *Root) MerkleRoot(h HashFn) Root {
	if s == nil {
		return Root{}
	}
	return *s
}

func sha256Combi(a Root, b Root) Root {
	v := [64]byte{}
	copy(v[:32], a[:])
	copy(v[32:], b[:])
	return sha256.Sum256(v[:])
}

var Hash HashFn = sha256Combi

var ZeroHashes []Root

// initialize the zero-hashes pre-computed data with the given hash-function.
func InitZeroHashes(h HashFn, zeroHashesLevels uint) {
	ZeroHashes = make([]Root, zeroHashesLevels+1)
	for i := uint(0); i < zeroHashesLevels; i++ {
		ZeroHashes[i+1] = h(ZeroHashes[i], ZeroHashes[i])
	}
}

func init() {
	InitZeroHashes(sha256Combi, 64)
}
