package operations

import (
	"context"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/protolambda/zrnt/eth2/beacon/bellatrix"
	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/zrnt/tests/spec/test_util"
)

type MockExecEngine struct {
	Valid bool `yaml:"execution_valid"`
}

func (m *MockExecEngine) ExecutePayload(ctx context.Context, executionPayload interface{}) (valid bool, err error) {
	return m.Valid, nil
}

var _ common.ExecutionEngine = (*MockExecEngine)(nil)

type ExecutionPayloadTestCase struct {
	test_util.BaseTransitionTest
	ExecutionPayload common.SpecObj
	Execution        MockExecEngine
}

func (c *ExecutionPayloadTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	switch forkName {
	case "bellatrix":
		c.ExecutionPayload = new(bellatrix.ExecutionPayload)
	case "capella":
		c.ExecutionPayload = new(capella.ExecutionPayload)
	case "eip4844":
		c.ExecutionPayload = new(deneb.ExecutionPayload)
	}
	test_util.LoadSSZ(t, "execution_payload", c.Spec.Wrap(c.ExecutionPayload), readPart)
	part := readPart.Part("execution.yml")
	dec := yaml.NewDecoder(part)
	dec.KnownFields(true)
	test_util.Check(t, dec.Decode(&c.Execution))
}

func (c *ExecutionPayloadTestCase) Run() error {
	switch s := c.Pre.(type) {
	case bellatrix.ExecutionTrackingBeaconState:
		return bellatrix.ProcessExecutionPayload(context.Background(), c.Spec, s, c.ExecutionPayload.(*bellatrix.ExecutionPayload), &c.Execution)
	case capella.ExecutionTrackingBeaconState:
		return capella.ProcessExecutionPayload(context.Background(), c.Spec, s, c.ExecutionPayload.(*capella.ExecutionPayload), &c.Execution)
	case deneb.ExecutionTrackingBeaconState:
		return deneb.ProcessExecutionPayload(context.Background(), c.Spec, s, c.ExecutionPayload.(*deneb.ExecutionPayload), &c.Execution)
	default:
		return fmt.Errorf("unrecognized state type: %T", c.Pre)
	}
}

func TestExecutionPayload(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"bellatrix", "capella", "eip4844"}, "operations", "execution_payload",
		func() test_util.TransitionTest { return new(ExecutionPayloadTestCase) })
}
