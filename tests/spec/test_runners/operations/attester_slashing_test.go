package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type AttesterSlashingTestCase struct {
	test_util.BaseTransitionTest
	AttesterSlashing beacon.AttesterSlashing
}

func (c *AttesterSlashingTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	test_util.LoadSpecObj(t, "attester_slashing", &c.AttesterSlashing, readPart)
}

func (c *AttesterSlashingTestCase) Run() error {
	epc, err := c.Spec.NewEpochsContext(c.Pre)
	if err != nil {
		return err
	}
	return c.Spec.ProcessAttesterSlashing(epc, c.Pre, &c.AttesterSlashing)
}

func TestAttesterSlashing(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "attester_slashing",
		func() test_util.TransitionTest { return new(AttesterSlashingTestCase) })
}
