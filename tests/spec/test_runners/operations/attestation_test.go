package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type AttestationTestCase struct {
	test_util.BaseTransitionTest
	Attestation phase0.Attestation
}

func (c *AttestationTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	test_util.LoadSpecObj(t, "attestation", &c.Attestation, readPart)
}

func (c *AttestationTestCase) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	return phase0.ProcessAttestation(c.Spec, epc, c.Pre, &c.Attestation)
}

func TestAttestation(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "attestation",
		func() test_util.TransitionTest { return new(AttestationTestCase) })
}
