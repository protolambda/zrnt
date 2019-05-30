package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	. "github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type ProposerSlashingTestCase struct {
	ProposerSlashing     *beacon.ProposerSlashing
	StateTransitionTestBase `mapstructure:",squash"`
}

func (testCase *ProposerSlashingTestCase) Process() error {
	return block_processing.ProcessProposerSlashing(testCase.Pre, testCase.ProposerSlashing)
}

func (testCase *ProposerSlashingTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestProposerSlashing(t *testing.T) {
	RunSuitesInPath("operations/proposer_slashing/",
		func(raw interface{}) interface{} { return new(ProposerSlashingTestCase) }, t)
}
