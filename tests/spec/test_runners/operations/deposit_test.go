package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon/deposits"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type DepositTestCase struct {
	test_util.BaseTransitionTest
	Deposit *deposits.Deposit
}

func (c *DepositTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	c.LoadSSZ(t, "deposit", c.Deposit, deposits.DepositSSZ, readPart)
}

func (c *DepositTestCase) Run() error {
	state := c.Prepare()
	return state.ProcessDeposit(c.Deposit)
}

func TestDeposit(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "deposit",
		func() test_util.TransitionTest {return new(DepositTestCase)})
}
