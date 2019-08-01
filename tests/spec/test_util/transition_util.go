package test_util

import (
	"github.com/protolambda/messagediff"
	"github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zssz"
	"github.com/protolambda/zssz/types"
	"testing"
)

type BaseTransitionTest struct {
	Pre  phase0.BeaconState
	Post phase0.BeaconState
}

func (c *BaseTransitionTest) Prepare() *phase0.FullFeaturedState {
	state := phase0.NewFullFeaturedState(&c.Pre)
	state.LoadPrecomputedData()
	return state
}

func (c *BaseTransitionTest) LoadSSZ(t *testing.T, name string, dst interface{}, ssz types.SSZ, readPart TestPartReader) {
	p := readPart(name + ".ssz")
	size, err := p.Size()
	Check(t, err)
	Check(t, zssz.Decode(p, size, dst, ssz))
	Check(t, p.Close())
}

func (c *BaseTransitionTest) Load(t *testing.T, readPart TestPartReader) {
	c.LoadSSZ(t, "pre", &c.Pre, phase0.BeaconStateSSZ, readPart)
	c.LoadSSZ(t, "post", &c.Post, phase0.BeaconStateSSZ, readPart)
}

func (c *BaseTransitionTest) Check(t *testing.T) {
	if diff, equal := messagediff.PrettyDiff(c.Pre, c.Post, messagediff.SliceWeakEmptyOption{}); !equal {
		t.Fatalf("end result does not match expectation!\n%s", diff)
	}
}

type TransitionTest interface {
	Load(t *testing.T, readPart TestPartReader)
	Run() error
	Check(t *testing.T)
}

type TransitionCaseMaker func() TransitionTest

func RunTransitionTest(t *testing.T, runnerName string, handlerName string, mkr TransitionCaseMaker) {
	RunHandler(t, runnerName + "/"+handlerName,
		HandleBLS(func(t *testing.T, readPart TestPartReader) {
			c := mkr()
			c.Load(t, readPart)
			if err := c.Run(); err != nil {
				t.Errorf("%s/%s process error: %v", runnerName, handlerName, err)
			}
			c.Check(t)
		}), core.PRESET_NAME)
}
