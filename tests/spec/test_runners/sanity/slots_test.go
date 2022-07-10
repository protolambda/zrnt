package sanity

import (
	"context"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v3"
)

type SlotsTestCase struct {
	test_util.BaseTransitionTest
	Slots common.Slot
}

func (c *SlotsTestCase) Load(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	p := readPart.Part("slots.yaml")
	dec := yaml.NewDecoder(p)
	test_util.Check(t, dec.Decode(&c.Slots))
	test_util.Check(t, p.Close())
}

type nonUpgradeable struct {
	common.BeaconState
}

func (*nonUpgradeable) UpgradeMaybe(ctx context.Context, spec *common.Spec, epc *common.EpochsContext) error {
	return nil
}

func (c *SlotsTestCase) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	slot, err := c.Pre.Slot()
	if err != nil {
		return err
	}
	return common.ProcessSlots(context.Background(), c.Spec, epc, &nonUpgradeable{c.Pre}, slot+c.Slots)
}

func TestSlots(t *testing.T) {
	test_util.RunTransitionTest(t, test_util.AllForks, "sanity", "slots",
		func() test_util.TransitionTest { return new(SlotsTestCase) })
}
