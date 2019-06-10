package sanity

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/transition"
	. "github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type BlocksTestCase struct {
	Blocks                  []*beacon.BeaconBlock
	StateTransitionTestBase `mapstructure:",squash"`
}

func (testCase *BlocksTestCase) Process() error {
	for _, block := range testCase.Blocks {
		// TODO: mark block number in error? (Maybe with Go 1.13, coming out soon, supporting wrapped errors)
		if err := transition.StateTransition(testCase.Pre, block, false); err != nil {
			return err
		}
	}
	return nil
}

func (testCase *BlocksTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestBlocks(t *testing.T) {
	RunSuitesInPath("sanity/blocks/",
		func(raw interface{}) (interface{}, interface{}) { return new(BlocksTestCase), raw }, t)
}
