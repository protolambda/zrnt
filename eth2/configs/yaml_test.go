package configs

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func mustLoad(name string, phase string) []byte {
	b, err := ioutil.ReadFile(filepath.Join("yamls", name, phase+".yaml"))
	if err != nil {
		panic(err)
	}
	return b
}

func TestYamlDecodingMainnetPhase0(t *testing.T) {
	var conf common.Phase0Config
	if err := yaml.Unmarshal(mustLoad("mainnet", "phase0"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.Phase0Config) {
		t.Fatal("Failed to load mainnet phase0 config")
	}
}

func TestYamlDecodingMainnetAltair(t *testing.T) {
	var conf common.AltairConfig
	if err := yaml.Unmarshal(mustLoad("mainnet", "altair"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.AltairConfig) {
		t.Fatal("Failed to load mainnet altair config")
	}
}

func TestYamlDecodingMainnetMerge(t *testing.T) {
	var conf common.MergeConfig
	if err := yaml.Unmarshal(mustLoad("mainnet", "merge"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.MergeConfig) {
		t.Fatal("Failed to load mainnet merge config")
	}
}

func TestYamlDecodingMainnetSharding(t *testing.T) {
	var conf common.ShardingConfig
	if err := yaml.Unmarshal(mustLoad("mainnet", "sharding"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.ShardingConfig) {
		t.Fatal("Failed to load mainnet sharding config")
	}
}

func TestYamlDecodingMinimalPhase0(t *testing.T) {
	var conf common.Phase0Config
	if err := yaml.Unmarshal(mustLoad("minimal", "phase0"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.Phase0Config) {
		t.Fatal("Failed to load minimal phase0 config")
	}
}

func TestYamlDecodingMinimalAltair(t *testing.T) {
	var conf common.AltairConfig
	if err := yaml.Unmarshal(mustLoad("minimal", "altair"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.AltairConfig) {
		t.Fatal("Failed to load minimal altair config")
	}
}

func TestYamlDecodingMinimalMerge(t *testing.T) {
	var conf common.MergeConfig
	if err := yaml.Unmarshal(mustLoad("minimal", "merge"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.MergeConfig) {
		t.Fatal("Failed to load minimal merge config")
	}
}

func TestYamlDecodingMinimalSharding(t *testing.T) {
	var conf common.ShardingConfig
	if err := yaml.Unmarshal(mustLoad("minimal", "sharding"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.ShardingConfig) {
		t.Fatal("Failed to load minimal sharding config")
	}
}
