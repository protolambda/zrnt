package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon/epoch_processing"
	. "github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type RegistryUpdatesTestCase struct {
	StateTransitionTestBase `mapstructure:",squash"`
}

func (testCase *RegistryUpdatesTestCase) Process() error {
	epoch_processing.ProcessEpochRegistryUpdates(testCase.Pre)
	return nil
}

func (testCase *RegistryUpdatesTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestRegistryUpdates(t *testing.T) {
	RunSuitesInPath("epoch_processing/registry_updates/",
		func(raw interface{}) interface{} { return new(RegistryUpdatesTestCase) }, t)
}
