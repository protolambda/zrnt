package test_util

import (
	"bytes"
	"github.com/protolambda/messagediff"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zssz"
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

func loadState(t *testing.T, name string, readPart TestPartReader) *beacon.BeaconStateView {
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

	if pre := loadState(t, "pre", readPart); pre != nil {
		c.Pre = pre
	} else {
		t.Fatalf("failed to load pre state")
	}
	if post := loadState(t, "post", readPart); post != nil {
		c.Post = post
	}
	// post state is optional, no error if not present.
}

func stateTreeToStateStruct(v *beacon.BeaconStateView) (*beacon.BeaconState, error) {
	var buf bytes.Buffer
	if err := v.Serialize(&buf); err != nil {
		return nil, err
	}
	var state beacon.BeaconState
	err := zssz.Decode(bytes.NewReader(buf.Bytes()), uint64(len(buf.Bytes())), &state, beacon.BeaconStateSSZ)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (c *BaseTransitionTest) Check(t *testing.T) {
	if c.ExpectingFailure() {
		t.Errorf("was expecting failure, but no error was raised")
	} else {
		hFn := tree.GetHashFn()
		preRoot := c.Pre.HashTreeRoot(hFn)
		postRoot := c.Post.HashTreeRoot(hFn)
		if preRoot != postRoot {
			// Hack to get the structural state representation, and then diff those.
			pre, err := stateTreeToStateStruct(c.Pre)
			if err != nil {
				t.Fatal(err)
			}
			post, err := stateTreeToStateStruct(c.Post)
			if err != nil {
				t.Fatal(err)
			}
			if diff, equal := messagediff.PrettyDiff(pre, post, messagediff.SliceWeakEmptyOption{}); !equal {
				t.Errorf("end result does not match expectation!\n%s", diff)
			}
		}
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
		}), beacon.PRESET_NAME)
}
