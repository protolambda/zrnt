package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon/slashings/attslash"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type AttesterSlashingTestCase struct {
	test_util.BaseTransitionTest
	AttesterSlashing *attslash.AttesterSlashing
}

func (c *AttesterSlashingTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	c.LoadSSZ(t, "attester_slashing", c.AttesterSlashing, attslash.AttesterSlashingSSZ, readPart)
}

func (c *AttesterSlashingTestCase) Run() error {
	state := c.Prepare()
	return state.ProcessAttesterSlashing(c.AttesterSlashing)
}

func TestAttesterSlashing(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "attester_slashing",
		func() test_util.TransitionTest {return new(AttesterSlashingTestCase)})
}
