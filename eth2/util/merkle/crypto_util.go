package merkle

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/math"
)

// MerkleRoot merkleizes values (where len(values) is a power of two) and returns the Merkle root.
// Note that the leaves are not hashed.
func MerkleRoot(values []Root) Root {
	if len(values) == 0 {
		return Root{}
	}
	if len(values) == 1 {
		return values[0]
	}
	power2 := math.NextPowerOfTwo(uint64(len(values)))
	o := make([]Root, power2<<1)
	copy(o[power2:], values)
	for i := int64(power2) - 1; i > 0; i-- {
		o[i] = hashing.Hash(append(o[i<<1][:], o[(i<<1)+1][:]...))
	}
	return o[1]
}

// VerifyMerkleBranch verifies that the given leaf is
// on the merkle branch at the given depth, at the index at that depth.
func VerifyMerkleBranch(leaf Root, branch []Root, depth uint64, index uint64, root Root) bool {
	value := leaf
	for i := uint64(0); i < depth; i++ {
		if (index>>i)&1 == 1 {
			value = hashing.Hash(append(branch[i][:], value[:]...))
		} else {
			value = hashing.Hash(append(value[:], branch[i][:]...))
		}
	}
	return value == root
}

// ConstructProof builds a merkle-branch of the given depth,
// as a proof of inclusion of the leaf (or something in the path to the root, with a smaller depth)
// at the given index (at that depth), for a list of leafs of a balanced binary hash-root-tree.
func ConstructProof(values []Root, index uint64, depth uint8) (branch []Root) {
	power2 := math.NextPowerOfTwo(uint64(len(values)))
	branch = make([]Root, depth)
	if power2 <= 1 {
		return branch
	}
	o := make([]Root, power2<<1)
	copy(o[power2:], values)
	for i := int64(power2) - 1; i > 0; i-- {
		o[i] = hashing.Hash(append(o[i<<1][:], o[(i<<1)+1][:]...))
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
