package merkle

import (
	"go-beacon-transition/eth2"
	"go-beacon-transition/eth2/util/hash"
)

// Merkleize values (where len(values) is a power of two) and return the Merkle root.
// Note that the leaves are not hashed.
func Merkle_root(values []eth2.Bytes32) eth2.Root {
	o := make([]eth2.Bytes32, len(values)*2)
	copy(o[len(values):], values)
	for i := len(values) - 1; i >= 0; i-- {
		o[i] = hash.Hash(append(o[i*2][:], o[i*2+1][:]...))
	}
	return eth2.Root(o[1])
}

// Verify that the given leaf is on the merkle branch.
func Verify_merkle_branch(leaf eth2.Bytes32, branch []eth2.Root, depth uint64, index uint64, root eth2.Root) bool {
	value := leaf
	for i := uint64(0); i < depth; i++ {
		if (index>>i)&1 == 1 {
			value = hash.Hash(append(branch[i][:], value[:]...))
		} else {
			value = hash.Hash(append(value[:], branch[i][:]...))
		}
	}
	return eth2.Root(value) == root
}
