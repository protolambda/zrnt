package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	"github.com/protolambda/zrnt/spec_testing"
	"testing"
)

type ProposerSlashingTestCase struct {
	ProposerSlashing     *beacon.ProposerSlashing
	OperationsTestBase `mapstructure:",squash"`
}

func (testCase *ProposerSlashingTestCase) Process() error {
	return block_processing.ProcessProposerSlashing(testCase.Pre, testCase.ProposerSlashing)
}

func (testCase *ProposerSlashingTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestProposerSlashing(t *testing.T) {
	spec_testing.RunSuitesInPath("../../../eth2.0-specs/eth2.0-spec-tests/tests/operations/proposer_slashing/",
		func(raw interface{}) interface{} { return new(ProposerSlashingTestCase) }, t)
}
