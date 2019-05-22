package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	"github.com/protolambda/zrnt/spec_testing"
	"testing"
)

type TransferTestCase struct {
	Transfer     *beacon.Transfer
	OperationsTestBase `mapstructure:",squash"`
}

func (testCase *TransferTestCase) Process() error {
	return block_processing.ProcessTransfer(testCase.Pre, testCase.Transfer)
}

func (testCase *TransferTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestTransfer(t *testing.T) {
	spec_testing.RunSuitesInPath("../../../eth2.0-specs/eth2.0-spec-tests/tests/operations/transfer/",
		func(raw interface{}) interface{} { return new(TransferTestCase) }, t)
}
