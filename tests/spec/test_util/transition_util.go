package test_util

import (
	"bytes"
	"fmt"
	"github.com/golang/snappy"
	"github.com/protolambda/messagediff"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/merge"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"io/ioutil"
	"testing"
)

// Fork where the test is organized, and thus the state/block/etc. types default to.
type ForkName string

var AllForks = []ForkName{"phase0", "altair", "merge"}

type BaseTransitionTest struct {
	Spec *common.Spec
	Fork ForkName
	Pre  common.BeaconState
	Post common.BeaconState
}

func (c *BaseTransitionTest) ExpectingFailure() bool {
	return c.Post == nil
}

func LoadState(t *testing.T, fork ForkName, name string, readPart TestPartReader) common.BeaconState {
	p := readPart.Part(name + ".ssz_snappy")
	spec := readPart.Spec()
	if p.Exists() {
		data, err := ioutil.ReadAll(p)
		Check(t, err)
		Check(t, p.Close())
		uncompressed, err := snappy.Decode(nil, data)
		Check(t, err)
		decodingReader := codec.NewDecodingReader(bytes.NewReader(uncompressed), uint64(len(uncompressed)))
		var state common.BeaconState
		switch fork {
		case "phase0":
			state, err = phase0.AsBeaconStateView(phase0.BeaconStateType(spec).Deserialize(decodingReader))
		case "altair":
			state, err = altair.AsBeaconStateView(altair.BeaconStateType(spec).Deserialize(decodingReader))
		case "merge":
			state, err = merge.AsBeaconStateView(merge.BeaconStateType(spec).Deserialize(decodingReader))
		default:
			t.Fatalf("unrecognized fork name: %s", fork)
			return nil
		}
		Check(t, err)
		return state
	} else {
		return nil
	}
}

func (c *BaseTransitionTest) Load(t *testing.T, forkName ForkName, readPart TestPartReader) {
	c.Spec = readPart.Spec()
	c.Fork = forkName
	if pre := LoadState(t, c.Fork, "pre", readPart); pre != nil {
		c.Pre = pre
	} else {
		t.Fatalf("failed to load pre state")
	}
	if post := LoadState(t, c.Fork, "post", readPart); post != nil {
		c.Post = post
	}
	// post state is optional, no error if not present.
}

func (c *BaseTransitionTest) Check(t *testing.T) {
	if c.ExpectingFailure() {
		t.Errorf("was expecting failure, but no error was raised")
	} else {
		diff, err := CompareStates(c.Spec, c.Pre, c.Post)
		if err != nil {
			t.Fatal(err)
		}
		if diff != "" {
			t.Errorf("end result does not match expectation!\n%s", diff)
		}
	}
}

func encodeStateForDiff(spec *common.Spec, state common.BeaconState) (interface{}, error) {
	switch s := state.(type) {
	case *phase0.BeaconStateView:
		return s.Raw(spec)
	case *altair.BeaconStateView:
		return s.Raw(spec)
	case *merge.BeaconStateView:
		return s.Raw(spec)
	default:
		return nil, fmt.Errorf("unrecognized beacon state type: %T", s)
	}
}

func CompareStates(spec *common.Spec, a common.BeaconState, b common.BeaconState) (diff string, err error) {
	hFn := tree.GetHashFn()
	preRoot := a.HashTreeRoot(hFn)
	postRoot := b.HashTreeRoot(hFn)
	if preRoot != postRoot {
		// Hack to get the structural state representation, and then diff those.
		pre, err := encodeStateForDiff(spec, a)
		if err != nil {
			return "", err
		}
		post, err := encodeStateForDiff(spec, b)
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
	Load(t *testing.T, forkName ForkName, readPart TestPartReader)
	ExpectingFailure() bool
	Run() error
	Check(t *testing.T)
}

type TransitionCaseMaker func() TransitionTest

func RunTransitionTest(t *testing.T, forks []ForkName, runnerName string, handlerName string, mkr TransitionCaseMaker) {
	caseRunner := HandleBLS(func(t *testing.T, forkName ForkName, readPart TestPartReader) {
		c := mkr()
		c.Load(t, forkName, readPart)
		if err := c.Run(); err != nil {
			if c.ExpectingFailure() {
				return
			}
			t.Fatalf("%s/%s process error: %v", runnerName, handlerName, err)
		}
		c.Check(t)
	})
	t.Run("minimal", func(t *testing.T) {
		for _, fork := range forks {
			t.Run(string(fork), func(t *testing.T) {
				RunHandler(t, runnerName+"/"+handlerName, caseRunner, configs.Minimal, fork)
			})
		}
	})
	t.Run("mainnet", func(t *testing.T) {
		for _, fork := range forks {
			t.Run(string(fork), func(t *testing.T) {
				RunHandler(t, runnerName+"/"+handlerName, caseRunner, configs.Mainnet, fork)
			})
		}
	})
}
