package test_util

import (
	"bytes"
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/execution"
	"io/ioutil"
	"testing"

	"github.com/golang/snappy"
	"github.com/protolambda/messagediff"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"gopkg.in/yaml.v3"

	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/bellatrix"
	"github.com/protolambda/zrnt/eth2/beacon/capella"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/deneb"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
)

// Fork where the test is organized, and thus the state/block/etc. types default to.
type ForkName string

var AllForks = []ForkName{"phase0", "altair", "bellatrix", "capella", "deneb"}

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
		case "bellatrix":
			state, err = bellatrix.AsBeaconStateView(bellatrix.BeaconStateType(spec).Deserialize(decodingReader))
		case "capella":
			state, err = capella.AsBeaconStateView(capella.BeaconStateType(spec).Deserialize(decodingReader))
		case "deneb":
			state, err = deneb.AsBeaconStateView(deneb.BeaconStateType(spec).Deserialize(decodingReader))
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

type BlocksTestCase struct {
	BaseTransitionTest
	Blocks []*common.BeaconBlockEnvelope
}

type BlocksCountMeta struct {
	BlocksCount uint64 `yaml:"blocks_count"`
}

func (c *BlocksTestCase) Load(t *testing.T, forkName ForkName, readPart TestPartReader) {
	c.BaseTransitionTest.Load(t, forkName, readPart)
	p := readPart.Part("meta.yaml")
	dec := yaml.NewDecoder(p)
	m := &BlocksCountMeta{}
	Check(t, dec.Decode(&m))
	Check(t, p.Close())
	valRoot, err := c.Pre.GenesisValidatorsRoot()
	if err != nil {
		t.Fatalf("failed to get pre-state genesis validators root: %v", err)
	}
	loadBlock := func(i uint64) *common.BeaconBlockEnvelope {
		switch forkName {
		case "phase0":
			dst := new(phase0.SignedBeaconBlock)
			LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.GENESIS_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		case "altair":
			dst := new(altair.SignedBeaconBlock)
			LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.ALTAIR_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		case "bellatrix":
			dst := new(bellatrix.SignedBeaconBlock)
			LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.BELLATRIX_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		case "capella":
			dst := new(capella.SignedBeaconBlock)
			LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.CAPELLA_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		case "deneb":
			dst := new(deneb.SignedBeaconBlock)
			LoadSpecObj(t, fmt.Sprintf("blocks_%d", i), dst, readPart)
			digest := common.ComputeForkDigest(c.Spec.DENEB_FORK_VERSION, valRoot)
			return dst.Envelope(c.Spec, digest)
		default:
			t.Fatalf("unrecognized fork name: %s", forkName)
			return nil
		}
	}
	for i := uint64(0); i < m.BlocksCount; i++ {
		c.Blocks = append(c.Blocks, loadBlock(i))
	}
}

func (c *BlocksTestCase) Run() error {
	epc, err := common.NewEpochsContext(c.Spec, c.Pre)
	if err != nil {
		return err
	}
	state := &beacon.StandardUpgradeableBeaconState{BeaconState: c.Pre}
	defer func() {
		c.Pre = state.BeaconState
	}()
	for _, b := range c.Blocks {
		if err := common.StateTransition(context.Background(), c.Spec, epc, state, b, true); err != nil {
			return err
		}
	}
	return nil
}

func encodeStateForDiff(spec *common.Spec, state common.BeaconState) (interface{}, error) {
	switch s := state.(type) {
	case *phase0.BeaconStateView:
		return s.Raw(spec)
	case *altair.BeaconStateView:
		return s.Raw(spec)
	case *bellatrix.BeaconStateView:
		return s.Raw(spec)
	case *capella.BeaconStateView:
		return s.Raw(spec)
	case *deneb.BeaconStateView:
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
			t.Errorf("%s/%s process error: %v", runnerName, handlerName, err)
		}
		c.Check(t)
	})
	t.Run("minimal", func(t *testing.T) {
		spec := *configs.Minimal
		spec.ExecutionEngine = &execution.NoOpExecutionEngine{}
		for _, fork := range forks {
			t.Run(string(fork), func(t *testing.T) {
				RunHandler(t, runnerName+"/"+handlerName, caseRunner, &spec, fork)
			})
		}
	})
	t.Run("mainnet", func(t *testing.T) {
		spec := *configs.Mainnet
		spec.ExecutionEngine = &execution.NoOpExecutionEngine{}
		for _, fork := range forks {
			t.Run(string(fork), func(t *testing.T) {
				RunHandler(t, runnerName+"/"+handlerName, caseRunner, &spec, fork)
			})
		}
	})
}
