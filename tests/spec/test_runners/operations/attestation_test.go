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
	test_util.LoadSSZ(t, "attestation", &c.Attestation, beacon.AttestationSSZ, readPart)
}

func (c *AttestationTestCase) Run() error {
	epc, err := c.Pre.NewEpochsContext()
	if err != nil {
		return err
	}
	return c.Pre.ProcessAttestation(epc, &c.Attestation)
}

func TestAttestation(t *testing.T) {
	test_util.RunTransitionTest(t, "operations", "attestation",
		func() test_util.TransitionTest { return new(AttestationTestCase) })
}
