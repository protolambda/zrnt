package sanity

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"

	"github.com/golang/snappy"
	"github.com/protolambda/ztyp/codec"
	"gopkg.in/yaml.v3"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/bellatrix"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/zrnt/tests/spec/test_util"
)

func runCase(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	spec := readPart.Spec()
	p := readPart.Part("genesis.ssz_snappy")
	data, err := ioutil.ReadAll(p)
	test_util.Check(t, err)
	test_util.Check(t, p.Close())
	uncompressed, err := snappy.Decode(nil, data)
	test_util.Check(t, err)
	decodingReader := codec.NewDecodingReader(bytes.NewReader(uncompressed), uint64(len(uncompressed)))
	var genesisState common.BeaconState
	switch forkName {
	case "phase0":
		genesisState, err = phase0.AsBeaconStateView(phase0.BeaconStateType(spec).Deserialize(decodingReader))
	case "altair":
		genesisState, err = altair.AsBeaconStateView(altair.BeaconStateType(spec).Deserialize(decodingReader))
	case "bellatrix":
		genesisState, err = bellatrix.AsBeaconStateView(bellatrix.BeaconStateType(spec).Deserialize(decodingReader))
	case "capella":
		genesisState, err = capella.AsBeaconStateView(capella.BeaconStateType(spec).Deserialize(decodingReader))
	case "deneb":
		genesisState, err = deneb.AsBeaconStateView(deneb.BeaconStateType(spec).Deserialize(decodingReader))
	default:
		t.Fatalf("unrecognized fork name: %s", forkName)
	}
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
}

func validity(spec *common.Spec) func(t *testing.T) {
	return func(t *testing.T) {
		for _, fork := range test_util.AllForks {
			t.Run(string(fork), func(t *testing.T) {
				test_util.RunHandler(t, "genesis/validity", runCase, spec, fork)
			})
		}
	}
}

func TestValidity(t *testing.T) {
	t.Run("minimal", validity(configs.Minimal))
	t.Run("mainnet", validity(configs.Mainnet))
}
