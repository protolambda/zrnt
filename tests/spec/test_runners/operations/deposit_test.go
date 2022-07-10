package operations

import (
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type DepositTestCase struct {
	test_util.BaseTransitionTest
	Deposit common.Deposit
}

func (c *DepositTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	test_util.LoadSSZ(t, "deposit", &c.Deposit, readPart)
}

func (c *DepositTestCase) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	return phase0.ProcessDeposit(c.Spec, epc, c.Pre, &c.Deposit, false)
}

func TestDeposit(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "operations", "deposit",
		func() test_util.TransitionTest { return new(DepositTestCase) })
}
