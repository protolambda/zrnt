package sanity

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestValidity(t *testing.T) {
	test_util.RunHandler(t, "genesis/validity",
		func(t *testing.T, readPart test_util.TestPartReader) {
			p := readPart("genesis.ssz")
			stateSize, err := p.Size()
			test_util.Check(t, err)
			genesisState, err := AsBeaconStateView(BeaconStateType.Deserialize(p, stateSize))
			test_util.Check(t, err)
			var expectedValid bool
			p = readPart("is_valid.yaml")
			dec := yaml.NewDecoder(p)
			test_util.Check(t, dec.Decode(&expectedValid))
			test_util.Check(t, p.Close())
			computedValid, err := IsValidGenesisState(genesisState)
			test_util.Check(t, err)
			if computedValid {
				if !expectedValid {
					t.Errorf("genesis state validity false positive")
				}
			} else {
				if expectedValid {
					t.Errorf("genesis state validity false negative")
				}
			}
		}, PRESET_NAME)
}
