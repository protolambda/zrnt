package shuffling

import (
	"fmt"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v3"
)

type ShufflingTestCase struct {
	Spec    *common.Spec            `yaml:"-"`
	Seed    common.Root             `yaml:"seed"`
	Count   uint64                  `yaml:"count"`
	Mapping []common.ValidatorIndex `yaml:"mapping"`
}

func (testCase *ShufflingTestCase) Run(t *testing.T) {
	t.Run(fmt.Sprintf("shuffle_%x_%d", testCase.Seed, testCase.Count), func(t *testing.T) {
		if testCase.Count != uint64(len(testCase.Mapping)) {
			t.Fatalf("invalid shuffling test")
		}
		t.Run("UnshuffleList", func(t *testing.T) {
			data := make([]common.ValidatorIndex, len(testCase.Mapping), len(testCase.Mapping))
			for i := 0; i < len(data); i++ {
				data[i] = common.ValidatorIndex(i)
			}
			common.UnshuffleList(uint8(testCase.Spec.SHUFFLE_ROUND_COUNT), data, testCase.Seed)
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
			data := make([]common.ValidatorIndex, len(testCase.Mapping), len(testCase.Mapping))
			for i := 0; i < len(data); i++ {
				data[i] = common.ValidatorIndex(i)
			}
			common.ShuffleList(uint8(testCase.Spec.SHUFFLE_ROUND_COUNT), data, testCase.Seed)
			for i := uint64(0); i < testCase.Count; i++ {
				shuffleOut := testCase.Mapping[i]
				shuffledIndex := data[shuffleOut]
				expectedIndex := common.ValidatorIndex(i)
				if shuffledIndex != expectedIndex {
					t.Errorf("different shuffled index: %d, expected %d, at index %d", shuffledIndex, expectedIndex, i)
					break
				}
			}
		})
		t.Run("UnpermuteIndex", func(t *testing.T) {
			for i := uint64(0); i < testCase.Count; i++ {
				shuffledIndex := testCase.Mapping[i]
				unshuffledIndex := common.UnpermuteIndex(uint8(testCase.Spec.SHUFFLE_ROUND_COUNT), shuffledIndex, testCase.Count, testCase.Seed)
				if unshuffledIndex != common.ValidatorIndex(i) {
					t.Errorf("different un-permuted index: %d (at %d) unshuffled to %d", shuffledIndex, i, unshuffledIndex)
					break
				}
			}
		})
		t.Run("PermuteIndex", func(t *testing.T) {
			for i := uint64(0); i < testCase.Count; i++ {
				expectedIndex := testCase.Mapping[i]
				shuffledIndex := common.PermuteIndex(uint8(testCase.Spec.SHUFFLE_ROUND_COUNT), common.ValidatorIndex(i), testCase.Count, testCase.Seed)
				if shuffledIndex != expectedIndex {
					t.Errorf("different shuffled index: %d, expected %d, at index %d", shuffledIndex, expectedIndex, i)
					break
				}
			}
		})
	})
}

func TestShuffling(t *testing.T) {
	runSpecShuffling := func(spec *common.Spec) func(t *testing.T) {
		return func(t *testing.T) {
			test_util.RunHandler(t, "shuffling/core/",
				func(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
					p := readPart.Part("mapping.yaml")
					dec := yaml.NewDecoder(p)
					c := &ShufflingTestCase{}
					test_util.Check(t, dec.Decode(&c))
					c.Spec = readPart.Spec()
					test_util.Check(t, p.Close())
					c.Run(t)
				}, configs.Mainnet, "phase0")
		}
	}
	t.Run("minimal", runSpecShuffling(configs.Minimal))
	t.Run("mainnet", runSpecShuffling(configs.Mainnet))
}
