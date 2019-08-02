package sanity

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestValidity(t *testing.T) {
	test_util.RunHandler(t, "genesis/validity",
		func(t *testing.T, readPart test_util.TestPartReader) {
			var genesisState phase0.BeaconState
			if !test_util.LoadSSZ(t, "genesis", &genesisState, phase0.BeaconStateSSZ, readPart) {
				t.Fatalf("no state to check genesis validity for")
			}
			var valid bool
			p := readPart("is_valid.yaml")
			dec := yaml.NewDecoder(p)
			test_util.Check(t, dec.Decode(&valid))
			test_util.Check(t, p.Close())
			if phase0.IsValidGenesisState(&genesisState) {
				if !valid {
					t.Errorf("genesis state validity false positive")
				}
			} else {
				if valid {
					t.Errorf("genesis state validity false negative")
				}
			}
		}, PRESET_NAME)
}
