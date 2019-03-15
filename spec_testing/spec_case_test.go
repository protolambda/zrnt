package spec_testing

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/transition"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"os"
	"testing"
)

type CaseData struct {

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

func TestSpecCase(t *testing.T) {

	// first argument being the code
	if len(os.Args) <= 1 {
		t.Errorf("test needs encoded case argument")
	}

	// last argument is assumed to be the base64 of the JSON encoded case data, with Go names encoding
	encodedCaseData := os.Args[len(os.Args) - 1]
	jsonBytes, err := base64.StdEncoding.DecodeString(encodedCaseData)
	if err != nil {
		t.Fatalf("cannot decode base64 input, %s", encodedCaseData)
	}
	data := CaseData{}

	t.Log(string(encodedCaseData))
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		t.Fatalf("cannot read spec test case data: %v", err)
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
