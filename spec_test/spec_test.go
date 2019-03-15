package spec_test

import (
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/transition"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
)

type SpecTestCase struct {

	// (constants are ignored, they are handled on compile time of the test)
	VerifySignatures bool `yaml:"verify_signatures"`
	// key/value pairs of fields of BeaconState
	InitialState *beacon.BeaconState `yaml:"initial_state"`
	// list of blocks
	Blocks []*beacon.BeaconBlock `yaml:"blocks"`
	// key/value pairs of subset of BeaconState fields to state
	ExpectedState *beacon.BeaconState `yaml:"expected_state"`
	// tree root
	ExpectedStateRoot beacon.Root `yaml:"expected_state_root"`
}

func BlockProcessTest(t *testing.T) {

	if len(os.Args) != 0 {
		t.Errorf("test needs config path argument")
	}

	configPath := os.Args[1]

	fileBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Errorf("cannot read config %s %v", configPath, err)
	}

	data := SpecTestCase{}

	if err := yaml.Unmarshal(fileBytes, &data); err != nil {
		t.Errorf("cannot read spec test case yaml: %v", err)
	}

	state := data.InitialState
	var transitionErr error
	for i, block := range data.Blocks {
		t.Run(fmt.Sprintf("process_block_%d", i), func(it *testing.T) {
			state, transitionErr = transition.StateTransition(state, block, false)
			if transitionErr != nil {
				it.Errorf("failed to do transition, triggered by block: %v", transitionErr)
			}
		})
	}

	t.Run("check_end_state", func(it *testing.T) {
		root := ssz.HashTreeRoot(state)
		if root != data.ExpectedStateRoot {
			// TODO: print diff
			t.Errorf("end result hashes do not match! Expected: %s, Got: %s", root, data.ExpectedStateRoot)
		}
	})
}
