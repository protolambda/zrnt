package operations

import (
	"context"
	"fmt"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/bellatrix"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v3"
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
	ExecutionPayload common.ExecutionPayload
	Execution        MockExecEngine
}

func (c *ExecutionPayloadTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	test_util.LoadSSZ(t, "sync_aggregate", c.Spec.Wrap(&c.ExecutionPayload), readPart)
	part := readPart.Part("execution.yml")
	dec := yaml.NewDecoder(part)
	dec.KnownFields(true)
	test_util.Check(t, dec.Decode(&c.Execution))
}

func (c *ExecutionPayloadTestCase) Run() error {
	s, ok := c.Pre.(*bellatrix.BeaconStateView)
	if !ok {
		return fmt.Errorf("unrecognized state type: %T", c.Pre)
	}
	return bellatrix.ProcessExecutionPayload(context.Background(), c.Spec, s, &c.ExecutionPayload, &c.Execution)
}

func TestExecutionPayload(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"bellatrix"}, "operations", "execution_payload",
		func() test_util.TransitionTest { return new(ExecutionPayloadTestCase) })
}
