package operations

import (
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type AttesterSlashingTestCase struct {
	test_util.BaseTransitionTest
	AttesterSlashing phase0.AttesterSlashing
}

func (c *AttesterSlashingTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	test_util.LoadSpecObj(t, "attester_slashing", &c.AttesterSlashing, readPart)
}

func (c *AttesterSlashingTestCase) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	return phase0.ProcessAttesterSlashing(c.Spec, epc, c.Pre, &c.AttesterSlashing)
}

func TestAttesterSlashing(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "operations", "attester_slashing",
		func() test_util.TransitionTest { return new(AttesterSlashingTestCase) })
}
