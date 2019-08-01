package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon/slashings/propslash"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type ProposerSlashingTestCase struct {
	test_util.BaseTransitionTest
	ProposerSlashing propslash.ProposerSlashing
}

func (c *ProposerSlashingTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	c.LoadSSZ(t, "proposer_slashing", &c.ProposerSlashing, propslash.ProposerSlashingSSZ, readPart)
}

func (c *ProposerSlashingTestCase) Run() error {
	state := c.Prepare()
	return state.ProcessProposerSlashing(&c.ProposerSlashing)
}

func TestProposerSlashing(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "proposer_slashing",
		func() test_util.TransitionTest {return new(ProposerSlashingTestCase)})
}
