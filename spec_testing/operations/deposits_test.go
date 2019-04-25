package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	"github.com/protolambda/zrnt/spec_testing"
	"gopkg.in/d4l3k/messagediff.v1"
	"testing"
)

type DepositTestCase struct {
	Description string
	Pre         *beacon.BeaconState
	Deposit     *beacon.Deposit
	Post        *beacon.BeaconState
}

func (testCase *DepositTestCase) Run(t *testing.T) {
	err := block_processing.ProcessDeposit(testCase.Pre, testCase.Deposit)
	if testCase.Post == nil {
		if err != nil {
			// expected error, test passes
			return
		} else {
			t.Fatalf("deposit should have thrown an error: %s", testCase.Description)
		}
	}

	if err != nil {
		t.Fatalf("deposit processing unexpectedly threw an error: %v", err)
	}

	// in case hashes are incorrectly correct (e.g. new SSZ behavior), we still have diffs
	if diff, equal := messagediff.PrettyDiff(testCase.Pre, testCase.Post); !equal {
		t.Fatalf("end result does not match expectation!\n%s", diff)
	}
}

func TestDeposits(t *testing.T) {
	spec_testing.RunSuitesInPath("../../../eth2.0-specs/eth2.0-spec-tests/tests/operations/deposits/",
		func(raw interface{}) interface{} { return new(DepositTestCase) }, t)
}
