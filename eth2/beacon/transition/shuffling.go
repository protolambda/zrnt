package transition

import (
	"encoding/binary"
	"go-beacon-transition/eth2"
	"go-beacon-transition/eth2/beacon"
	"go-beacon-transition/eth2/util/hash"
)

const hSeedSize, hRoundSize, hPositionWindowSize = int8(32), int8(1), int8(4)
const hPivotViewSize = hSeedSize + hRoundSize
const hTotalSize = hSeedSize + hRoundSize + hPositionWindowSize


// NOTE: this is going to be replaced with an imported protolambda/eth2-shuffle.
// No need to codegolf it for spec anymore.


// Adapted from other project of @protolambda: https://github.com/protolambda/eth2-shuffle
// (Tested there, this is almost a 1-1 copy)
func shuffleValidatorIndices(input []eth2.ValidatorIndex, seed eth2.Bytes32) {
	if len(input) <= 1 {
		// nothing to (un)shuffle
		return
	}
	listSize := uint64(len(input))
	buf := make([]byte, hTotalSize, hTotalSize)
	// Seed is always the first 32 bytes of the hash input, we never have to change this part of the buffer.
	copy(buf[:hSeedSize], seed[:])
	for r := uint8(0); r < eth2.SHUFFLE_ROUND_COUNT; r++ {
		// spec: pivot = bytes_to_int(hash(seed + int_to_bytes1(round))[0:8]) % list_size
		// This is the "int_to_bytes1(round)", appended to the seed.
		buf[hSeedSize] = r
		// Seed is already in place, now just hash the correct part of the buffer, and take a uint64 from it,
		//  and modulo it to get a pivot within range.
		h := hash.Hash(buf[:hPivotViewSize])
		pivot := binary.LittleEndian.Uint64(h[:8]) % listSize

		// Split up the for-loop in two:
		//  1. Handle the part from 0 (incl) to pivot (incl). This is mirrored around (pivot / 2)
		//  2. Handle the part from pivot (excl) to N (excl). This is mirrored around ((pivot / 2) + (size/2))
		// The pivot defines a split in the array, with each of the splits mirroring their data within the split.
		// Print out some example even/odd sized index lists, with some even/odd pivots,
		//  and you can deduce how the mirroring works exactly.
		// Note that the mirror is strict enough to not consider swapping the index @mirror with itself.
		mirror := (pivot + 1) >> 1
		// Since we are iterating through the "positions" in order, we can just repeat the hash every 256th position.
		// No need to pre-compute every possible hash for efficiency like in the example code.
		// We only need it consecutively (we are going through each in reverse order however, but same thing)
		//
		// spec: source = hash(seed + int_to_bytes1(round) + int_to_bytes4(position // 256))
		// - seed is still in 0:32 (excl., 32 bytes)
		// - round number is still in 32
		// - mix in the position for randomness, except the last byte of it,
		//     which will be used later to select a bit from the resulting hash.
		// We start from the pivot position, and work back to the mirror position (of the part left to the pivot).
		// This makes us process each pear exactly once (instead of unnecessarily twice, like in the spec)
		binary.LittleEndian.PutUint32(buf[hPivotViewSize:], uint32(pivot>>8))
		source := hash.Hash(buf)
		byteV := source[(pivot&0xff)>>3]
		handle := func(i uint64, j uint64) {
			// The pair is i,j. With j being the bigger of the two, hence the "position" identifier of the pair.
			// Every 256th bit (aligned to j).
			if j&0xff == 0xff {
				// just overwrite the last part of the buffer, reuse the start (seed, round)
				binary.LittleEndian.PutUint32(buf[hPivotViewSize:], uint32(j>>8))
				source = hash.Hash(buf)
			}
			// Same trick with byte retrieval. Only every 8th.
			if j&0x7 == 0x7 {
				byteV = source[(j&0xff)>>3]
			}

			if (byteV>>(j&0x7))&0x1 == 1 {
				// swap the pair items
				input[i], input[j] = input[j], input[i]
			}
		}
		for i, j := uint64(0), pivot; i < mirror; i, j = i+1, j-1 {
			handle(i, j)
		}
		// Now repeat, but for the part after the pivot.
		mirror = (pivot + listSize + 1) >> 1
		end := listSize - 1
		// Again, seed and round input is in place, just update the position.
		// We start at the end, and work back to the mirror point.
		// This makes us process each pear exactly once (instead of unnecessarily twice, like in the spec)
		binary.LittleEndian.PutUint32(buf[hPivotViewSize:], uint32(end>>8))
		source = hash.Hash(buf)
		byteV = source[(end&0xff)>>3]
		for i, j := pivot+1, end; i < mirror; i, j = i+1, j-1 {
			// Exact same thing
			handle(i, j)
		}
	}
}
