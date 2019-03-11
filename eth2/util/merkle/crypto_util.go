package merkle

import (
	"github.com/protolambda/zrnt/eth2/util/hash"
	"github.com/protolambda/zrnt/eth2/util/math"
)

// Merkleize values (where len(values) is a power of two) and return the Merkle root.
// Note that the leaves are not hashed.
func MerkleRoot(values [][32]byte) [32]byte {
	if len(values) == 0 {
		return [32]byte{}
	}
	if len(values) == 1 {
		return values[0]
	}
	power2 := math.NextPowerOfTwo(uint64(len(values)))
	o := make([][32]byte, power2<<1)
	copy(o[power2:], values)
	for i := int64(power2) - 1; i > 0; i-- {
		o[i] = hash.Hash(append(o[i<<1][:], o[(i<<1)+1][:]...))
	}
	return o[1]
}

// Verify that the given leaf is on the merkle branch.
func VerifyMerkleBranch(leaf [32]byte, branch [][32]byte, depth uint64, index uint64, root [32]byte) bool {
	value := leaf
	for i := uint64(0); i < depth; i++ {
		if (index>>i)&1 == 1 {
			value = hash.Hash(append(branch[i][:], value[:]...))
		} else {
			value = hash.Hash(append(value[:], branch[i][:]...))
		}
	}
	return value == root
}

func ConstructProof(values [][32]byte, index uint64, depth uint8) (branch [][32]byte) {
	power2 := math.NextPowerOfTwo(uint64(len(values)))
	branch = make([][32]byte, depth)
	if power2 <= 1 {
		return branch
	}
	o := make([][32]byte, power2<<1)
	copy(o[power2:], values)
	for i := int64(power2) - 1; i > 0; i-- {
		o[i] = hash.Hash(append(o[i<<1][:], o[(i<<1)+1][:]...))
	}
	depthOffset := power2 << 1
	for j := uint8(0); j < depth; j++ {
		depthOffset >>= 1
		if depthOffset <= 1 {
			break
		}
		indexAtDepth := index >> j
		branchIndexAtDepth := indexAtDepth ^ 1
		branch[j] = o[depthOffset+branchIndexAtDepth]
	}
	return branch
}
