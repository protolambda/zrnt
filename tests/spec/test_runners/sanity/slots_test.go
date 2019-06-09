package sanity

import (
	"github.com/protolambda/zrnt/eth2/beacon/transition"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type SlotsTestCase struct {
	Slots                   Slot
	StateTransitionTestBase `mapstructure:",squash"`
}

func (testCase *SlotsTestCase) Process() error {
	// TODO: mark block number in error? (Maybe with Go 1.13, coming out soon, supporting wrapped errors)
	if err := transition.StateTransitionTo(testCase.Pre, testCase.Pre.Slot+testCase.Slots); err != nil {
		return err
	}
	return nil
}

func (testCase *SlotsTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestSlots(t *testing.T) {
	RunSuitesInPath("sanity/slots/",
		func(raw interface{}) (interface{}, interface{}) { return new(SlotsTestCase), raw }, t)
}
