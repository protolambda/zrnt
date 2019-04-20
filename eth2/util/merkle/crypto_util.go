package merkle

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/hash"
	"github.com/protolambda/zrnt/eth2/util/math"
)

// Merkleize values (where len(values) is a power of two) and return the Merkle root.
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
		o[i] = hash.HashRoot(append(o[i<<1][:], o[(i<<1)+1][:]...))
	}
	return o[1]
}

// Verify that the given leaf is on the merkle branch.
func VerifyMerkleBranch(leaf Root, branch []Root, depth uint64, index uint64, root Root) bool {
	value := leaf
	for i := uint64(0); i < depth; i++ {
		if (index>>i)&1 == 1 {
			value = hash.HashRoot(append(branch[i][:], value[:]...))
		} else {
			value = hash.HashRoot(append(value[:], branch[i][:]...))
		}
	}
	return value == root
}

func ConstructProof(values []Root, index uint64, depth uint8) (branch []Root) {
	power2 := math.NextPowerOfTwo(uint64(len(values)))
	branch = make([]Root, depth)
	if power2 <= 1 {
		return branch
	}
	o := make([]Root, power2<<1)
	copy(o[power2:], values)
	for i := int64(power2) - 1; i > 0; i-- {
		o[i] = hash.HashRoot(append(o[i<<1][:], o[(i<<1)+1][:]...))
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
