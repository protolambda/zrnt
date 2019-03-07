package merkle

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/util/hash"
	"github.com/protolambda/go-beacon-transition/eth2/util/math"
)

// Merkleize values (where len(values) is a power of two) and return the Merkle root.
// Note that the leaves are not hashed.
func Merkle_root(values []beacon.Bytes32) beacon.Root {
	if len(values) == 0 {
		return beacon.Root{}
	}
	if len(values) == 1 {
		return beacon.Root(values[0])
	}
	power2 := math.NextPowerOfTwo(uint64(len(values)))
	o := make([]beacon.Bytes32, power2<<1)
	copy(o[power2:], values)
	for i := int64(power2) - 1; i >= 0; i-- {
		o[i] = hash.Hash(append(o[i<<1][:], o[(i<<1)+1][:]...))
	}
	return beacon.Root(o[1])
}

// Verify that the given leaf is on the merkle branch.
func Verify_merkle_branch(leaf beacon.Bytes32, branch []beacon.Root, depth uint64, index uint64, root beacon.Root) bool {
	value := leaf
	for i := uint64(0); i < depth; i++ {
		if (index>>i)&1 == 1 {
			value = hash.Hash(append(branch[i][:], value[:]...))
		} else {
			value = hash.Hash(append(value[:], branch[i][:]...))
		}
	}
	return beacon.Root(value) == root
}
