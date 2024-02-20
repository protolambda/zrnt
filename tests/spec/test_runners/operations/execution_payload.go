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

func (m *MockExecEngine) DenebNotifyNewPayload(ctx context.Context, executionPayload *deneb.ExecutionPayload, parentBeaconBlockRoot common.Root) (valid bool, err error) {
	return m.Valid, nil
}

func (m *MockExecEngine) DenebIsValidVersionedHashes(ctx context.Context, payload *deneb.ExecutionPayload, versionedHashes []common.Hash32) (bool, error) {
	return m.Valid, nil
}

func (m *MockExecEngine) DenebIsValidBlockHash(ctx context.Context, payload *deneb.ExecutionPayload, parentBeaconBlockRoot common.Root) (bool, error) {
	return m.Valid, nil
}

func (m *MockExecEngine) CapellaNotifyNewPayload(ctx context.Context, executionPayload *capella.ExecutionPayload) (valid bool, err error) {
	return m.Valid, nil
}

func (m *MockExecEngine) CapellaIsValidBlockHash(ctx context.Context, payload *capella.ExecutionPayload) (bool, error) {
	return m.Valid, nil
}

func (m *MockExecEngine) BellatrixNotifyNewPayload(ctx context.Context, executionPayload *bellatrix.ExecutionPayload) (valid bool, err error) {
	return m.Valid, nil
}

func (m *MockExecEngine) BellatrixIsValidBlockHash(ctx context.Context, payload *bellatrix.ExecutionPayload) (bool, error) {
	return m.Valid, nil
}

var _ bellatrix.ExecutionEngine = (*MockExecEngine)(nil)
var _ capella.ExecutionEngine = (*MockExecEngine)(nil)
var _ deneb.ExecutionEngine = (*MockExecEngine)(nil)

func (m *MockExecEngine) ExecutePayload(ctx context.Context, executionPayload interface{}) (valid bool, err error) {
	return m.Valid, nil
}

var _ common.ExecutionEngine = (*MockExecEngine)(nil)

type ExecutionPayloadTestCase struct {
	test_util.BaseTransitionTest
	BlockBody common.SpecObj
	Execution MockExecEngine
}

func (c *ExecutionPayloadTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	switch forkName {
	case "bellatrix":
		c.BlockBody = new(bellatrix.BeaconBlockBody)
	case "capella":
		c.BlockBody = new(capella.BeaconBlockBody)
	case "deneb":
		c.BlockBody = new(deneb.BeaconBlockBody)
	}
	test_util.LoadSSZ(t, "body", c.Spec.Wrap(c.BlockBody), readPart)
	part := readPart.Part("execution.yml")
	dec := yaml.NewDecoder(part)
	dec.KnownFields(true)
	test_util.Check(t, dec.Decode(&c.Execution))
}

func (c *ExecutionPayloadTestCase) Run() error {
	switch s := c.Pre.(type) {
	case bellatrix.ExecutionTrackingBeaconState:
		return bellatrix.ProcessExecutionPayload(context.Background(), c.Spec,
			s, &c.BlockBody.(*bellatrix.BeaconBlockBody).ExecutionPayload, &c.Execution)
	case capella.ExecutionTrackingBeaconState:
		return capella.ProcessExecutionPayload(context.Background(), c.Spec,
			s, &c.BlockBody.(*capella.BeaconBlockBody).ExecutionPayload, &c.Execution)
	case deneb.ExecutionTrackingBeaconState:
		return deneb.ProcessExecutionPayload(context.Background(), c.Spec,
			s, c.BlockBody.(*deneb.BeaconBlockBody), &c.Execution)
	default:
		return fmt.Errorf("unrecognized state type: %T", c.Pre)
	}
}

func TestExecutionPayload(t *testing.T) {
	test_util.RunTransitionTest(t, []test_util.ForkName{"bellatrix", "capella", "deneb"}, "operations", "execution_payload",
		func() test_util.TransitionTest { return new(ExecutionPayloadTestCase) })
}
