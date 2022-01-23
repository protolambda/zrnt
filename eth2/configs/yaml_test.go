package configs

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
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

func TestYamlDecodingMainnetSharding(t *testing.T) {
	var conf common.ShardingPreset
	if err := yaml.Unmarshal(mustLoad("presets", "mainnet", "sharding"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.ShardingPreset) {
		t.Fatal("Failed to load mainnet sharding preset")
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

func TestYamlDecodingMinimalSharding(t *testing.T) {
	var conf common.ShardingPreset
	if err := yaml.Unmarshal(mustLoad("presets", "minimal", "sharding"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.ShardingPreset) {
		t.Fatal("Failed to load minimal sharding preset")
	}
}
