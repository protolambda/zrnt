package test_util

import (
	"github.com/protolambda/messagediff"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/ztyp/tree"
	"testing"
)

type BaseTransitionTest struct {
	Pre  *beacon.BeaconStateView
	Post *beacon.BeaconStateView
}

func (c *BaseTransitionTest) ExpectingFailure() bool {
	return c.Post == nil
}

func LoadState(t *testing.T, name string, readPart TestPartReader) *beacon.BeaconStateView {
	p := readPart(name + ".ssz")
	if p.Exists() {
		size, err := p.Size()
		Check(t, err)
		state, err := beacon.AsBeaconStateView(beacon.BeaconStateType.Deserialize(p, size))
		Check(t, err)
		Check(t, p.Close())
		return state
	} else {
		return nil
	}
}
func (c *BaseTransitionTest) Load(t *testing.T, readPart TestPartReader) {

	if pre := LoadState(t, "pre", readPart); pre != nil {
		c.Pre = pre
	} else {
		t.Fatalf("failed to load pre state")
	}
	if post := LoadState(t, "post", readPart); post != nil {
		c.Post = post
	}
	// post state is optional, no error if not present.
}

func (c *BaseTransitionTest) Check(t *testing.T) {
	if c.ExpectingFailure() {
		t.Errorf("was expecting failure, but no error was raised")
	} else {
		diff, err := CompareStates(c.Pre, c.Post)
		if err != nil {
			t.Fatal(err)
		}
		if diff != "" {
			t.Errorf("end result does not match expectation!\n%s", diff)
		}
	}
}

func CompareStates(a *beacon.BeaconStateView, b *beacon.BeaconStateView) (diff string, err error) {
	hFn := tree.GetHashFn()
	preRoot := a.HashTreeRoot(hFn)
	postRoot := b.HashTreeRoot(hFn)
	if preRoot != postRoot {
		// Hack to get the structural state representation, and then diff those.
		pre, err := a.Raw()
		if err != nil {
			return "", err
		}
		post, err := b.Raw()
		if err != nil {
			return "", err
		}
		if diff, equal := messagediff.PrettyDiff(pre, post, messagediff.SliceWeakEmptyOption{}); !equal {
			return diff, nil
		}
	}
	return "", nil
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
		}), beacon.PRESET_NAME)
}
