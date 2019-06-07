package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon/epoch_processing"
	. "github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type CrosslinksTestCase struct {
	StateTransitionTestBase `mapstructure:",squash"`
}

func (testCase *CrosslinksTestCase) Process() error {
	epoch_processing.ProcessEpochCrosslinks(testCase.Pre)
	return nil
}

func (testCase *CrosslinksTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestCrosslinks(t *testing.T) {
	RunSuitesInPath("epoch_processing/crosslinks/",
		func(raw interface{}) (interface{}, interface {}) { return new(CrosslinksTestCase), raw }, t)
}
