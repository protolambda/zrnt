package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon/transfers"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type TransferTestCase struct {
	test_util.BaseTransitionTest
	Transfer transfers.Transfer
}

func (c *TransferTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	test_util.LoadSSZ(t, "transfer", &c.Transfer, transfers.TransferSSZ, readPart)
}

func (c *TransferTestCase) Run() error {
	state := c.Prepare()
	return state.ProcessTransfer(&c.Transfer)
}

func TestTransfer(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "transfer",
		func() test_util.TransitionTest { return new(TransferTestCase) })
}
