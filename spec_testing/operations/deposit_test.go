package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	"github.com/protolambda/zrnt/spec_testing"
	"testing"
)

type DepositTestCase struct {
	Deposit     *beacon.Deposit
	OperationsTestBase `mapstructure:",squash"`
}

func (testCase *DepositTestCase) Process() error {
	return block_processing.ProcessDeposit(testCase.Pre, testCase.Deposit)
}

func (testCase *DepositTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestDeposit(t *testing.T) {
	spec_testing.RunSuitesInPath("../../../eth2.0-specs/eth2.0-spec-tests/tests/operations/deposit/",
		func(raw interface{}) interface{} { return new(DepositTestCase) }, t)
}
