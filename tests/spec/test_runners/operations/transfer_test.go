package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	. "github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type TransferTestCase struct {
	Transfer     *beacon.Transfer
	StateTransitionTestBase `mapstructure:",squash"`
}

func (testCase *TransferTestCase) Process() error {
	return block_processing.ProcessTransfer(testCase.Pre, testCase.Transfer)
}

func (testCase *TransferTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestTransfer(t *testing.T) {
	RunSuitesInPath("operations/transfer/",
		func(raw interface{}) interface{} { return new(TransferTestCase) }, t)
}
