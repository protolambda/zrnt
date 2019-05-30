package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	. "github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type BlockHeaderTestCase struct {
	Block     *beacon.BeaconBlock
	StateTransitionTestBase `mapstructure:",squash"`
}

func (testCase *BlockHeaderTestCase) Process() error {
	return block_processing.ProcessBlockHeader(testCase.Pre, testCase.Block)
}

func (testCase *BlockHeaderTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestBlockHeader(t *testing.T) {
	RunSuitesInPath("operations/block_header/",
		func(raw interface{}) interface{} { return new(BlockHeaderTestCase) }, t)
}
