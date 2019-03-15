package spec_testing

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"gopkg.in/d4l3k/messagediff.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

type CaseData struct {
	// (constants are ignored, they are handled on compile time of the test)
	VerifySignatures bool
	// key/value pairs of fields of BeaconState
	InitialState *beacon.BeaconState
	// list of blocks
	Blocks []*beacon.BeaconBlock
	// key/value pairs of subset of BeaconState fields to state
	ExpectedState *beacon.BeaconState
	// tree root
	ExpectedStateRoot beacon.Root
}

func TestSpecCase(t *testing.T) {

	data, err := loadCaseData()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("g slot %d", beacon.GENESIS_SLOT)

	// Starting state
	state := data.InitialState
	//for i, block := range data.Blocks {
	//	// now run a sub-test: process the block
	//	t.Run(fmt.Sprintf("process_block_%d", i), func(it *testing.T) {
	//		var transitionErr error
	//		state, transitionErr = transition.StateTransition(state, block, false)
	//		if transitionErr != nil {
	//			it.Fatalf("failed to do transition, triggered by block: %v", transitionErr)
	//		}
	//	})
	//}

	// We processed every block successfully, now verify the end result
	t.Run("check_end_state", func(it *testing.T) {
		root := ssz.HashTreeRoot(state)
		if root != data.ExpectedStateRoot {
			t.Errorf("end result hashes do not match! Expected: %s, Got: %x", data.ExpectedStateRoot, root)
		}
		// in case hashes are incorrectly correct (e.g. new SSZ behavior), we still have diffs
		if diff, equal := messagediff.PrettyDiff(data.ExpectedState, state); !equal {
			t.Errorf("end result does not match expectation!\n%s", diff)
		}
	})
}

// Loads the test case from program arguments
func loadCaseData() (*CaseData, error) {

	// first argument being the code
	if len(os.Args) <= 2 {
		return nil, errors.New("test needs suite file and case index argument")
	}

	// last argument is assumed to be the base64 of the JSON encoded case data, with Go names encoding
	suiteFilePath := os.Args[len(os.Args)-2]
	var caseIndex uint32
	if v, err := strconv.ParseInt(os.Args[len(os.Args)-1], 10, 32); err != nil {
		return nil, errors.New("invalid test case index")
	} else {
		caseIndex = uint32(v)
	}

	yamlBytes, err := ioutil.ReadFile(suiteFilePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read test cases file %s %v", suiteFilePath, err))
	}

	suiteData := SpecTestsSuite{}
	if err := yaml.Unmarshal(yamlBytes, &suiteData); err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read spec test case data: %v", err))
	}

	if caseIndex > uint32(len(suiteData.TestCases)) {
		return nil, errors.New(fmt.Sprintf("case %d does not exist, only found %d test cases in suite %s", caseIndex, len(suiteData.TestCases), suiteFilePath))
	}
	caseData := suiteData.TestCases[caseIndex]

	decodedCaseData := CaseData{}
	// Hack: encode, and decode. Bit slow, but makes it easy to load the data into a fully typed struct
	encoded, err := json.Marshal(DecodeSpecFormat(caseData))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot parse spec data: %v", err))
	}
	if err := json.Unmarshal(encoded, &decodedCaseData); err != nil {
		return nil, errors.New(fmt.Sprintf("cannot decode test-case %s %d into case data struct: %v", suiteFilePath, caseIndex, err))
	}

	return &decodedCaseData, nil
}
