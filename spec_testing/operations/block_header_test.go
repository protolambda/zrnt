package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	"github.com/protolambda/zrnt/spec_testing"
	"testing"
)

type BlockHeaderTestCase struct {
	Block     *beacon.BeaconBlock
	OperationsTestBase `mapstructure:",squash"`
}

func (testCase *BlockHeaderTestCase) Process() error {
	return block_processing.ProcessBlockHeader(testCase.Pre, testCase.Block)
}

func (testCase *BlockHeaderTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestBlockHeader(t *testing.T) {
	spec_testing.RunSuitesInPath("../../../eth2.0-specs/eth2.0-spec-tests/tests/operations/block_header/",
		func(raw interface{}) interface{} { return new(BlockHeaderTestCase) }, t)
}
