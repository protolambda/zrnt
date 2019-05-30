package test_util

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"gopkg.in/d4l3k/messagediff.v1"
	"testing"
)

type StateTransitionTest interface {
	Process() error
	Base() *StateTransitionTestBase
}

type StateTransitionTestBase struct {
	Description string
	BlsSetting int
	Pre         *beacon.BeaconState
	Post        *beacon.BeaconState
}

const (
	BLS_OPTIONAL = 0
	BLS_REQUIRED = 1
	BLS_IGNORED = 2
)

func (base *StateTransitionTestBase) Title() string {
	return base.Description
}

func (base *StateTransitionTestBase) Base() *StateTransitionTestBase {
	return base
}

func RunTest(t *testing.T, testCase StateTransitionTest) {
	base := testCase.Base()
	if base.BlsSetting == BLS_REQUIRED {
		t.Log("skipping BLS-only test")
		return
	}
	err := testCase.Process()
	if base.Post == nil {
		if err != nil {
			// expected error, test passes
			return
		} else {
			t.Fatalf("operation should have thrown an error: %s", base.Description)
		}
	}

	if err != nil {
		t.Fatalf("operation processing unexpectedly threw an error: %v", err)
	}

	// in case hashes are incorrectly correct (e.g. new SSZ behavior), we still have diffs
	if diff, equal := messagediff.PrettyDiff(base.Pre, base.Post); !equal {
		t.Fatalf("end result does not match expectation!\n%s", diff)
	}
}
