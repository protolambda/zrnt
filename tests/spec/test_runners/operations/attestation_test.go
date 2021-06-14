package operations

import (
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"testing"
)

type AttestationTestCase struct {
	test_util.BaseTransitionTest
	Attestation phase0.Attestation
}

func (c *AttestationTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	test_util.LoadSpecObj(t, "attestation", &c.Attestation, readPart)
}

func (c *AttestationTestCase) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	if s, ok := c.Pre.(phase0.Phase0PendingAttestationsBeaconState); ok {
		return phase0.ProcessAttestation(c.Spec, epc, s, &c.Attestation)
	} else if s, ok := c.Pre.(*altair.BeaconStateView); ok {
		return altair.ProcessAttestation(c.Spec, epc, s, &c.Attestation)
	} else {
		return fmt.Errorf("unrecognized state type: %T", s)
	}
}

func TestAttestation(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "operations", "attestation",
		func() test_util.TransitionTest { return new(AttestationTestCase) })
}
