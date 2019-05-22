package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/block_processing"
	"github.com/protolambda/zrnt/spec_testing"
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
	spec_testing.RunSuitesInPath("../../../eth2.0-specs/eth2.0-spec-tests/tests/operations/attestation/",
		func(raw interface{}) interface{} { return new(AttestationTestCase) }, t)
}
