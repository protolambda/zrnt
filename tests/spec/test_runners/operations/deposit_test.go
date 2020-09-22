package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type DepositTestCase struct {
	test_util.BaseTransitionTest
	Deposit beacon.Deposit
}

func (c *DepositTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	test_util.LoadSSZ(t, "deposit", &c.Deposit, readPart)
}

func (c *DepositTestCase) Run() error {
	epc, err := c.Spec.NewEpochsContext(c.Pre)
	if err != nil {
		return err
	}
	return c.Spec.ProcessDeposit(epc, c.Pre, &c.Deposit, false)
}

func TestDeposit(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "deposit",
		func() test_util.TransitionTest { return new(DepositTestCase) })
}
