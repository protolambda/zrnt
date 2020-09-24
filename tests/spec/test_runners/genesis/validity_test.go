package sanity

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"github.com/protolambda/ztyp/codec"
	"gopkg.in/yaml.v3"
	"testing"
)

func validity(spec *Spec) func(t *testing.T) {
	return func(t *testing.T) {
		test_util.RunHandler(t, "genesis/validity",
			func(t *testing.T, readPart test_util.TestPartReader) {
				p := readPart.Part("genesis.ssz")
				stateSize, err := p.Size()
				test_util.Check(t, err)
				genesisState, err := AsBeaconStateView(spec.BeaconState().Deserialize(codec.NewDecodingReader(p, stateSize)))
				test_util.Check(t, err)
				var expectedValid bool
				p = readPart.Part("is_valid.yaml")
				dec := yaml.NewDecoder(p)
				test_util.Check(t, dec.Decode(&expectedValid))
				test_util.Check(t, p.Close())
				computedValid, err := spec.IsValidGenesisState(genesisState)
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
			}, spec)
	}
}

func TestValidity(t *testing.T) {
	t.Run("minimal", validity(configs.Minimal))
	t.Run("mainnet", validity(configs.Mainnet))
}
