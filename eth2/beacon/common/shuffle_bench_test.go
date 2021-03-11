package common

import (
	"fmt"
	"testing"
)

const benchShuffleRounds = 90

func BenchmarkPermuteIndex(b *testing.B) {
	listSizes := []uint64{4000000, 40000, 400}

	// "random" seed for testing. Can be any 32 bytes.
	seed := [32]byte{123, 42}

	for _, listSize := range listSizes {
		// benchmark!
		b.Run(fmt.Sprintf("PermuteIndex_%d", listSize), func(ib *testing.B) {
			for i := uint64(0); i < uint64(ib.N); i++ {
				PermuteIndex(benchShuffleRounds, ValidatorIndex(i%listSize), listSize, seed)
			}
		})
	}
}

func BenchmarkIndexComparison(b *testing.B) {
	// 4M is just too inefficient to even start comparing.
	listSizes := []uint64{40000, 400}

	// "random" seed for testing. Can be any 32 bytes.
	seed := [32]byte{123, 42}

	for _, listSize := range listSizes {
		// benchmark!
		b.Run(fmt.Sprintf("Indexwise_ShuffleList_%d", listSize), func(ib *testing.B) {
			for i := 0; i < ib.N; i++ {
				// Simulate a list-shuffle by running permute-index listSize times.
				for j := uint64(0); j < listSize; j++ {
					PermuteIndex(benchShuffleRounds, ValidatorIndex(j), listSize, seed)
				}
			}
		})
	}
}

func BenchmarkShuffleList(b *testing.B) {
	listSizes := []uint64{4000000, 40000, 400}

	// "random" seed for testing. Can be any 32 bytes.
	seed := [32]byte{123, 42}

	for _, listSize := range listSizes {
		// list to test
		testIndices := make([]ValidatorIndex, listSize, listSize)
		// fill
		for i := uint64(0); i < listSize; i++ {
			testIndices[i] = ValidatorIndex(i)
		}
		// benchmark!
		b.Run(fmt.Sprintf("ShuffleList_%d", listSize), func(ib *testing.B) {
			for i := 0; i < ib.N; i++ {
				ShuffleList(benchShuffleRounds, testIndices, seed)
			}
		})
	}
}
