package test_util

import (
	"github.com/protolambda/messagediff"
	"github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/phase0"
	"testing"
)

type BaseTransitionTest struct {
	Pre  *phase0.BeaconState
	Post *phase0.BeaconState
}

func (c *BaseTransitionTest) ExpectingFailure() bool {
	return c.Post == nil
}

func (c *BaseTransitionTest) Prepare() *phase0.FullFeaturedState {
	state := phase0.NewFullFeaturedState(c.Pre)
	state.LoadPrecomputedData()
	return state
}

func (c *BaseTransitionTest) Load(t *testing.T, readPart TestPartReader) {
	pre := new(phase0.BeaconState)
	if LoadSSZ(t, "pre", pre, phase0.BeaconStateSSZ, readPart) {
		c.Pre = pre
	} else {
		t.Fatalf("failed to load pre state")
	}
	post := new(phase0.BeaconState)
	if LoadSSZ(t, "post", post, phase0.BeaconStateSSZ, readPart) {
		c.Post = post
	}
	// post state is optional, no error if not present.
}

func (c *BaseTransitionTest) Check(t *testing.T) {
	if c.ExpectingFailure() {
		t.Errorf("was expecting failure, but no error was raised")
	} else if diff, equal := messagediff.PrettyDiff(c.Pre, c.Post, messagediff.SliceWeakEmptyOption{}); !equal {
		t.Errorf("end result does not match expectation!\n%s", diff)
	}
}

type TransitionTest interface {
	Load(t *testing.T, readPart TestPartReader)
	ExpectingFailure() bool
	Run() error
	Check(t *testing.T)
}

type TransitionCaseMaker func() TransitionTest

func RunTransitionTest(t *testing.T, runnerName string, handlerName string, mkr TransitionCaseMaker) {
	RunHandler(t, runnerName+"/"+handlerName,
		HandleBLS(func(t *testing.T, readPart TestPartReader) {
			c := mkr()
			c.Load(t, readPart)
			if err := c.Run(); err != nil {
				if c.ExpectingFailure() {
					return
				}
				t.Errorf("%s/%s process error: %v", runnerName, handlerName, err)
			}
			c.Check(t)
		}), core.PRESET_NAME)
}
