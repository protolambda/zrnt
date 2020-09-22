package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type VoluntaryExitTestCase struct {
	test_util.BaseTransitionTest
	VoluntaryExit beacon.SignedVoluntaryExit
}

func (c *VoluntaryExitTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	test_util.LoadSSZ(t, "voluntary_exit", &c.VoluntaryExit, readPart)
}

func (c *VoluntaryExitTestCase) Run() error {
	epc, err := c.Spec.NewEpochsContext(c.Pre)
	if err != nil {
		return err
	}
	return c.Spec.ProcessVoluntaryExit(epc, c.Pre, &c.VoluntaryExit)
}

func TestVoluntaryExit(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "voluntary_exit",
		func() test_util.TransitionTest { return new(VoluntaryExitTestCase) })
}
