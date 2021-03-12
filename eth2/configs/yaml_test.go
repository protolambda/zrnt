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

func TestYamlDecodingMainnetPhase1(t *testing.T) {
	var conf common.Phase1Config
	if err := yaml.Unmarshal(mustLoad("mainnet", "phase1"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Mainnet.Phase1Config) {
		t.Fatal("Failed to load mainnet phase1 config")
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

func TestYamlDecodingMinimalPhase1(t *testing.T) {
	var conf common.Phase1Config
	if err := yaml.Unmarshal(mustLoad("minimal", "phase1"), &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, Minimal.Phase1Config) {
		t.Fatal("Failed to load minimal phase1 config")
	}
}
