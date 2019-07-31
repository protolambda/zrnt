package epoch_processing

import (
	"github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

func TestRegistryUpdates(t *testing.T) {
	test_util.RunHandler(t, "epoch_processing/registry_updates/",
		MakeRunner(func(t *testing.T, testCase TestCase) {
			state := phase0.NewFullFeaturedState(testCase.Pre)
			state.LoadPrecomputedData()
			state.ProcessEpochRegistryUpdates()
			testCase.TestCondition(t)
		}), core.PRESET_NAME)
}
