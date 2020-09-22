package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type AttestationTestCase struct {
	test_util.BaseTransitionTest
	Attestation beacon.Attestation
}

func (c *AttestationTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	test_util.LoadSpecObj(t, "attestation", &c.Attestation, readPart)
}

func (c *AttestationTestCase) Run() error {
	epc, err := c.Spec.NewEpochsContext(c.Pre)
	if err != nil {
		return err
	}
	return c.Spec.ProcessAttestation(epc, c.Pre, &c.Attestation)
}

func TestAttestation(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "attestation",
		func() test_util.TransitionTest { return new(AttestationTestCase) })
}
