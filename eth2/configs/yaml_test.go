package configs

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
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

func TestYamlDecodingMainnetConfig(t *testing.T) {
	var conf common.Config
	if err := yaml.Unmarshal(mustLoad("configs", "mainnet"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.Config) {
		t.Fatal("Failed to load mainnet config")
	}
}

func TestYamlDecodingMinimalConfig(t *testing.T) {
	var conf common.Config
	if err := yaml.Unmarshal(mustLoad("configs", "minimal"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.Config) {
		t.Fatal("Failed to load minimal config")
	}
}

func TestYamlDecodingMainnetPhase0(t *testing.T) {
	var conf common.Phase0Preset
	if err := yaml.Unmarshal(mustLoad("presets", "mainnet", "phase0"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.Phase0Preset) {
		t.Fatal("Failed to load mainnet phase0 preset")
	}
}

func TestYamlDecodingMainnetAltair(t *testing.T) {
	var conf common.AltairPreset
	if err := yaml.Unmarshal(mustLoad("presets", "mainnet", "altair"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.AltairPreset) {
		t.Fatal("Failed to load mainnet altair preset")
	}
}

func TestYamlDecodingMainnetBellatrix(t *testing.T) {
	var conf common.BellatrixPreset
	if err := yaml.Unmarshal(mustLoad("presets", "mainnet", "bellatrix"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.BellatrixPreset) {
		t.Fatal("Failed to load mainnet bellatrix preset")
	}
}

func TestYamlDecodingMainnetCapella(t *testing.T) {
	var conf common.CapellaPreset
	if err := yaml.Unmarshal(mustLoad("presets", "mainnet", "capella"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.CapellaPreset) {
		t.Fatal("Failed to load mainnet capella preset")
	}
}

func TestYamlDecodingMainnetDeneb(t *testing.T) {
	var conf common.DenebPreset
	if err := yaml.Unmarshal(mustLoad("presets", "mainnet", "deneb"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.DenebPreset) {
		t.Fatal("Failed to load mainnet deneb preset")
	}
}

func TestYamlDecodingMinimalPhase0(t *testing.T) {
	var conf common.Phase0Preset
	if err := yaml.Unmarshal(mustLoad("presets", "minimal", "phase0"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.Phase0Preset) {
		t.Fatal("Failed to load minimal phase0 preset")
	}
}

func TestYamlDecodingMinimalAltair(t *testing.T) {
	var conf common.AltairPreset
	if err := yaml.Unmarshal(mustLoad("presets", "minimal", "altair"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.AltairPreset) {
		t.Fatal("Failed to load minimal altair preset")
	}
}

func TestYamlDecodingMinimalBellatrix(t *testing.T) {
	var conf common.BellatrixPreset
	if err := yaml.Unmarshal(mustLoad("presets", "minimal", "bellatrix"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.BellatrixPreset) {
		t.Fatal("Failed to load minimal bellatrix preset")
	}
}

func TestYamlDecodingMinimalCapella(t *testing.T) {
	var conf common.CapellaPreset
	if err := yaml.Unmarshal(mustLoad("presets", "minimal", "capella"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.CapellaPreset) {
		t.Fatal("Failed to load minimal capella preset")
	}
}

func TestYamlDecodingMinimalDeneb(t *testing.T) {
	var conf common.DenebPreset
	if err := yaml.Unmarshal(mustLoad("presets", "minimal", "deneb"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.DenebPreset) {
		t.Fatal("Failed to load minimal deneb preset")
	}
}
