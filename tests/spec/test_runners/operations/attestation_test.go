package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	"github.com/protolambda/zrnt/tests/spec/test_runners"
	"testing"
)

type AttestationTestCase struct {
	Attestation     *beacon.Attestation
	OperationsTestBase `mapstructure:",squash"`
}

func (testCase *AttestationTestCase) Process() error {
	return block_processing.ProcessAttestation(testCase.Pre, testCase.Attestation)
}

func (testCase *AttestationTestCase) Run(t *testing.T) {
	RunTest(t, testCase)
}

func TestAttestation(t *testing.T) {
	test_runners.RunSuitesInPath("operations/attestation/",
		func(raw interface{}) interface{} { return new(AttestationTestCase) }, t)
}
