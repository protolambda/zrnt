package sanity

import (
	"context"
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v3"
	"testing"
)

type SlotsTestCase struct {
	test_util.BaseTransitionTest
	Slots Slot
}

func (c *SlotsTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.BaseTransitionTest.Load(t, readPart)
	p := readPart.Part("slots.yaml")
	dec := yaml.NewDecoder(p)
	test_util.Check(t, dec.Decode(&c.Slots))
	test_util.Check(t, p.Close())
}

func (c *SlotsTestCase) Run() error {
	epc, err := c.Spec.NewEpochsContext(c.Pre)
	if err != nil {
		return err
	}
	slot, err := c.Pre.Slot()
	if err != nil {
		return err
	}
	return c.Spec.ProcessSlots(context.Background(), epc, c.Pre, slot+c.Slots)
}

func TestSlots(t *testing.T) {
	test_util.RunTransitionTest(t, "sanity", "slots",
		func() test_util.TransitionTest { return new(SlotsTestCase) })
}
