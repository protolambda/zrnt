package sanity

import (
	"bytes"
	"github.com/golang/snappy"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"github.com/protolambda/ztyp/codec"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"testing"
)

func validity(spec *common.Spec) func(t *testing.T) {
	return func(t *testing.T) {
		test_util.RunHandler(t, "genesis/validity",
			func(t *testing.T, readPart test_util.TestPartReader) {
				p := readPart.Part("genesis.ssz_snappy")
				data, err := ioutil.ReadAll(p)
				test_util.Check(t, err)
				test_util.Check(t, p.Close())
				uncompressed, err := snappy.Decode(nil, data)
				test_util.Check(t, err)
				genesisState, err := phase0.AsBeaconStateView(
					phase0.BeaconStateType(spec).Deserialize(codec.NewDecodingReader(
						bytes.NewReader(uncompressed), uint64(len(uncompressed)))))
				test_util.Check(t, err)
				var expectedValid bool
				p = readPart.Part("is_valid.yaml")
				dec := yaml.NewDecoder(p)
				test_util.Check(t, dec.Decode(&expectedValid))
				test_util.Check(t, p.Close())
				computedValid, err := phase0.IsValidGenesisState(spec, genesisState)
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
