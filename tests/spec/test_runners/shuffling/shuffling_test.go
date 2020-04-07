package operations

import (
	"encoding/hex"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v2"
	"testing"
)

type ShufflingTestCase struct {
	Seed    Root             `yaml:"-"`
	SeedRaw string           `yaml:"seed"`
	Count   uint64           `yaml:"count"`
	Mapping []ValidatorIndex `yaml:"mapping"`
}

func (testCase *ShufflingTestCase) Run(t *testing.T) {
	t.Run(fmt.Sprintf("shuffle_%x_%d", testCase.Seed, testCase.Count), func(t *testing.T) {
		if testCase.Count != uint64(len(testCase.Mapping)) {
			t.Fatalf("invalid shuffling test")
		}
		t.Run("UnshuffleList", func(t *testing.T) {
			data := make([]ValidatorIndex, len(testCase.Mapping), len(testCase.Mapping))
			for i := 0; i < len(data); i++ {
				data[i] = ValidatorIndex(i)
			}
			UnshuffleList(data, testCase.Seed)
			for i := uint64(0); i < testCase.Count; i++ {
				unshuffledIndex := data[i]
				expectedIndex := testCase.Mapping[i]
				if unshuffledIndex != expectedIndex {
					t.Errorf("different unshuffled index: %d at %d", unshuffledIndex, expectedIndex)
					break
				}
			}
		})
		t.Run("ShuffleList", func(t *testing.T) {
			data := make([]ValidatorIndex, len(testCase.Mapping), len(testCase.Mapping))
			for i := 0; i < len(data); i++ {
				data[i] = ValidatorIndex(i)
			}
			ShuffleList(data, testCase.Seed)
			for i := uint64(0); i < testCase.Count; i++ {
				shuffleOut := testCase.Mapping[i]
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
				shuffledIndex := testCase.Mapping[i]
				unshuffledIndex := UnpermuteIndex(shuffledIndex, testCase.Count, testCase.Seed)
				if unshuffledIndex != ValidatorIndex(i) {
					t.Errorf("different un-permuted index: %d (at %d) unshuffled to %d", shuffledIndex, i, unshuffledIndex)
					break
				}
			}
		})
		t.Run("PermuteIndex", func(t *testing.T) {
			for i := uint64(0); i < testCase.Count; i++ {
				expectedIndex := testCase.Mapping[i]
				shuffledIndex := PermuteIndex(ValidatorIndex(i), testCase.Count, testCase.Seed)
				if shuffledIndex != expectedIndex {
					t.Errorf("different shuffled index: %d, expected %d, at index %d", shuffledIndex, expectedIndex, i)
					break
				}
			}
		})
	})
}

func TestShuffling(t *testing.T) {
	test_util.RunHandler(t, "shuffling/core/", func(t *testing.T, readPart test_util.TestPartReader) {
		p := readPart("mapping.yaml")
		dec := yaml.NewDecoder(p)
		c := &ShufflingTestCase{}
		test_util.Check(t, dec.Decode(&c))
		test_util.Check(t, p.Close())
		seed, err := hex.DecodeString(c.SeedRaw[2:])
		test_util.Check(t, err)
		copy(c.Seed[:], seed)
		c.Run(t)
	}, PRESET_NAME)
}
