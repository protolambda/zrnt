package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	"github.com/protolambda/zrnt/spec_testing"
	"testing"
)

type VoluntaryExitTestCase struct {
	VoluntaryExit     *beacon.VoluntaryExit
	OperationsTestBase `mapstructure:",squash"`
}

func (testCase *VoluntaryExitTestCase) Process() error {
	return block_processing.ProcessVoluntaryExit(testCase.Pre, testCase.VoluntaryExit)
}

func (testCase *VoluntaryExitTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestVoluntaryExit(t *testing.T) {
	spec_testing.RunSuitesInPath("../../../eth2.0-specs/eth2.0-spec-tests/tests/operations/voluntary_exit/",
		func(raw interface{}) interface{} { return new(VoluntaryExitTestCase) }, t)
}
