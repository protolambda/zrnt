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

// Merkleize values (where len(values) is a power of two) and return the Merkle root.
// Values are supplied in a merkle-worksheet: twice the size of len(values), with in-between hash computations in the first half, and the values (padded) in the second half
// Note that the leaves are not hashed.
// The cached parameter tells for each value if it changed, compared to prev-work
func MerkleRootCached(prevWork [][32]byte, cached []bool) (root [32]byte, thisWork [][32]byte) {
	if len(cached) == 0 {
		return [32]byte{}, nil
	}
	if len(cached) == 1 {
		return prevWork[1], prevWork
	}
	power2 := math.NextPowerOfTwo(uint64(len(cached)))
	// our merkle worksheet, carries in-between hashing work, for re-use later on
	thisWork = make([][32]byte, power2<<1)
	if prevLen, thisLen := uint64(len(prevWork)), uint64(len(thisWork)); prevLen < thisLen {
		// previous work is smaller, we have to extend our merkle worksheet, just 2x difference would like this:
		// old: bbccccddddddddeeeeeeeeeeeeeeee                                  (lowercase is existing trie nodes)
		// how:   bb cccc     dddddddd        eeeeeeeeeeeeeeee
		// new: AAbbBBccccCCCCddddddddDDDDDDDDeeeeeeeeeeeeeeeeEEEEEEEEEEEEEEEE  (uppercase is new trie nodes)
		prevPower2 := math.NextPowerOfTwo(uint64(len(prevWork))) >> 1
		diffPower2 := power2 - prevPower2
		for i := uint64(1); i <= prevPower2; i++ {
			srcS := (1 << i)-2
			srcE := (2 << i)-2
			dstS := ((1 << i) << diffPower2)-2
			dstE := ((2 << i) << diffPower2)-2
			copy(thisWork[dstS:dstE], prevWork[srcS:srcE])
		}
	} else if prevLen > thisLen {
		// previous work is bigger, we can shrink our merkle worksheet, just 2x difference would like this:
		// old: bbccccddddddddeeeeeeeeeeeeeeee          (lowercase is existing trie nodes)
		// how:   cc  dddd    eeeeeeee
		// new: ccddddeeeeeeee                          (uppercase is new trie nodes)
		// effectively the reverse of the above extension of the merkle worksheet
		prevPower2 := math.NextPowerOfTwo(uint64(len(prevWork))) >> 1
		diffPower2 := prevPower2 - power2
		for i := uint64(1); i <= power2; i++ {
			srcS := ((1 << i) << diffPower2)-2
			srcE := ((2 << i) << diffPower2)-2
			dstS := (1 << i)-2
			dstE := (2 << i)-2
			copy(thisWork[dstS:dstE], prevWork[srcS:srcE])
		}
	} else {
		// ideal situation, merkle worksheet sizes match, we don't have to do much work to compute the new head
		copy(thisWork, prevWork)
	}

	// keep track of what changed in a mapping with the same structure as a worksheet.
	cachedWork := make([]bool, power2<<1)
	copy(cachedWork[power2:], cached)
	for i := int64(power2) - 1; i > 0; i-- {
		// if both source nodes are cached, then their combination does not have to be re-computed
		cachedWork[i] = cachedWork[i<<1] && cachedWork[(i<<1)+1]
		if !cachedWork[i] {
			thisWork[i] = hash.Hash(append(thisWork[i<<1][:], thisWork[(i<<1)+1][:]...))
		}
	}
	return thisWork[1], thisWork
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
