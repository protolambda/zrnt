package operations

import (
	"context"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/capella"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type BlsToExecutionChangeTestCase struct {
	test_util.BaseTransitionTest
	BlsToExecutionChange common.SignedBLSToExecutionChange
}

func (c *BlsToExecutionChangeTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	test_util.LoadSSZ(t, "address_change", &c.BlsToExecutionChange, readPart)
}

func (c *BlsToExecutionChangeTestCase) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	return capella.ProcessBLSToExecutionChange(context.Background(), c.Spec, epc, c.Pre, &c.BlsToExecutionChange)
}

func TestBlsToExecutionChange(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"capella", "deneb"}, "operations", "bls_to_execution_change",
		func() test_util.TransitionTest { return new(BlsToExecutionChangeTestCase) })
}
