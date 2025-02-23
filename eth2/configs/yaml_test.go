package configs

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/protolambda/zrnt/eth2/beacon/common"
)

func mustLoad(path ...string) []byte {
	b, err := ioutil.ReadFile(filepath.Join("yamls", filepath.Join(path...)) + ".yaml")
	if err != nil {
		panic(err)
	}
	return b
}

func yamlTest[E any](t *testing.T, path ...string) {
	t.Run("yaml-"+strings.Join(path, "-"), func(t *testing.T) {
		var conf E
		if err := yaml.Unmarshal(mustLoad(path...), &conf); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		roundTripTest(t, conf)
	})
}

func roundTripTest[E any](t *testing.T, conf E) {
	t.Run("roundtrip", func(t *testing.T) {
		data, err := yaml.Marshal(conf)
		if err != nil {
			t.Fatalf("failed to encode: %v", err)
		}
		var conf2 E
		if err := yaml.Unmarshal(data, &conf2); err != nil {
			t.Fatalf("failed to decode again: %v", err)
		}
		if !reflect.DeepEqual(conf, conf2) {
			t.Fatal("Failed to roundtrip")
		}
	})
}

func TestConfigs(t *testing.T) {
	for _, presetName := range []string{"minimal", "mainnet"} {
		yamlTest[common.Config](t, "configs", presetName)
		yamlTest[common.Phase0Preset](t, "presets", presetName, "phase0")
		yamlTest[common.AltairPreset](t, "presets", presetName, "altair")
		yamlTest[common.BellatrixPreset](t, "presets", presetName, "bellatrix")
		yamlTest[common.CapellaPreset](t, "presets", presetName, "capella")
		yamlTest[common.DenebPreset](t, "presets", presetName, "deneb")
		yamlTest[common.ElectraPreset](t, "presets", presetName, "electra")
	}
	roundTripTest[common.Spec](t, *Minimal)
	roundTripTest[common.Spec](t, *Mainnet)
}
