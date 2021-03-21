package test_util

import (
	"bytes"
	"github.com/golang/snappy"
	"github.com/protolambda/messagediff"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"io/ioutil"
	"testing"
)

type BaseTransitionTest struct {
	Spec *common.Spec
	Pre  *phase0.BeaconStateView
	Post *phase0.BeaconStateView
}

func (c *BaseTransitionTest) ExpectingFailure() bool {
	return c.Post == nil
}

func LoadState(t *testing.T, name string, readPart TestPartReader) *phase0.BeaconStateView {
	p := readPart.Part(name + ".ssz_snappy")
	spec := readPart.Spec()
	if p.Exists() {
		data, err := ioutil.ReadAll(p)
		Check(t, err)
		Check(t, p.Close())
		uncompressed, err := snappy.Decode(nil, data)
		Check(t, err)
		state, err := phase0.AsBeaconStateView(phase0.BeaconStateType(spec).Deserialize(
			codec.NewDecodingReader(bytes.NewReader(uncompressed), uint64(len(uncompressed)))))
		Check(t, err)
		return state
	} else {
		return nil
	}
}

func (c *BaseTransitionTest) Load(t *testing.T, readPart TestPartReader) {
	c.Spec = readPart.Spec()
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
		diff, err := CompareStates(c.Spec, c.Pre, c.Post)
		if err != nil {
			t.Fatal(err)
		}
		if diff != "" {
			t.Errorf("end result does not match expectation!\n%s", diff)
		}
	}
}

func CompareStates(spec *common.Spec, a *phase0.BeaconStateView, b *phase0.BeaconStateView) (diff string, err error) {
	hFn := tree.GetHashFn()
	preRoot := a.HashTreeRoot(hFn)
	postRoot := b.HashTreeRoot(hFn)
	if preRoot != postRoot {
		// Hack to get the structural state representation, and then diff those.
		pre, err := a.Raw(spec)
		if err != nil {
			return "", err
		}
		post, err := b.Raw(spec)
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
	caseRunner := HandleBLS(func(t *testing.T, readPart TestPartReader) {
		c := mkr()
		c.Load(t, readPart)
		if err := c.Run(); err != nil {
			if c.ExpectingFailure() {
				return
			}
			t.Errorf("%s/%s process error: %v", runnerName, handlerName, err)
		}
		c.Check(t)
	})
	t.Run("minimal", func(t *testing.T) {
		RunHandler(t, runnerName+"/"+handlerName, caseRunner, configs.Minimal)
	})
	t.Run("mainnet", func(t *testing.T) {
		RunHandler(t, runnerName+"/"+handlerName, caseRunner, configs.Mainnet)
	})
}
