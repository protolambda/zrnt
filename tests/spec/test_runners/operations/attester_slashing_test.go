package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	"github.com/protolambda/zrnt/tests/spec/test_runners"
	"testing"
)

type AttesterSlashingTestCase struct {
	AttesterSlashing     *beacon.AttesterSlashing
	OperationsTestBase `mapstructure:",squash"`
}

func (testCase *AttesterSlashingTestCase) Process() error {
	return block_processing.ProcessAttesterSlashing(testCase.Pre, testCase.AttesterSlashing)
}

func (testCase *AttesterSlashingTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestAttesterSlashing(t *testing.T) {
	test_runners.RunSuitesInPath("operations/attester_slashing/",
		func(raw interface{}) interface{} { return new(AttesterSlashingTestCase) }, t)
}
