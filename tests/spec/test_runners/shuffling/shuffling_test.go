package operations

import (
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/shuffling"
	. "github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type ShufflingTestCase struct {
	Seed Root
	Count uint64
	Shuffled []ValidatorIndex
}

func (testCase *ShufflingTestCase) Run(t *testing.T) {
	t.Run(fmt.Sprintf("shuffle_%x_%d", testCase.Seed, testCase.Count), func(t *testing.T) {
		if testCase.Count != uint64(len(testCase.Shuffled)) {
			t.Fatalf("invalid shuffling test")
		}
		t.Run("UnshuffleList", func(t *testing.T) {
			data := make([]ValidatorIndex, len(testCase.Shuffled), len(testCase.Shuffled))
			for i := 0; i < len(data); i++ {
				data[i] = ValidatorIndex(i)
			}
			shuffling.UnshuffleList(data, testCase.Seed)
			for i := uint64(0); i < testCase.Count; i++ {
				unshuffledIndex := data[i]
				expectedIndex := testCase.Shuffled[i]
				if unshuffledIndex != expectedIndex {
					t.Errorf("different unshuffled index: %d at %d", unshuffledIndex, expectedIndex)
					break
				}
			}
		})
		t.Run("ShuffleList", func(t *testing.T) {
			data := make([]ValidatorIndex, len(testCase.Shuffled), len(testCase.Shuffled))
			for i := 0; i < len(data); i++ {
				data[i] = ValidatorIndex(i)
			}
			shuffling.ShuffleList(data, testCase.Seed)
			for i := uint64(0); i < testCase.Count; i++ {
				shuffleOut := testCase.Shuffled[i]
				shuffledIndex := data[shuffleOut]
				expectedIndex := ValidatorIndex(i)
				if shuffledIndex != expectedIndex {
					t.Errorf("different shuffled index: %d, expected %d, at index %d", shuffledIndex, expectedIndex, i)
					break
				}
			}
		})
		t.Run("UnpermuteIndex", func(t *testing.T) {
			for i := uint64(0); i < testCase.Count; i++ {
				shuffledIndex := testCase.Shuffled[i]
				unshuffledIndex := shuffling.UnpermuteIndex(shuffledIndex, testCase.Count, testCase.Seed)
				if unshuffledIndex != ValidatorIndex(i) {
					t.Errorf("different un-permuted index: %d (at %d) unshuffled to %d", shuffledIndex, i, unshuffledIndex)
					break
				}
			}
		})
		t.Run("PermuteIndex", func(t *testing.T) {
			for i := uint64(0); i < testCase.Count; i++ {
				expectedIndex := testCase.Shuffled[i]
				shuffledIndex := shuffling.PermuteIndex(ValidatorIndex(i), testCase.Count, testCase.Seed)
				if shuffledIndex != expectedIndex {
					t.Errorf("different shuffled index: %d, expected %d, at index %d", shuffledIndex, expectedIndex, i)
					break
				}
			}
		})
	})
}

func TestShuffling(t *testing.T) {
	RunSuitesInPath("shuffling/core/",
		func(raw interface{}) (interface{}, interface{}) { return new(ShufflingTestCase), raw }, t)
}
