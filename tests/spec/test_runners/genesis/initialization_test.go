package sanity

import (
	"encoding/hex"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"github.com/protolambda/ztyp/codec"
	"gopkg.in/yaml.v3"
	"testing"
)

type InitializationTestCase struct {
	Spec          *Spec
	GenesisState  *BeaconStateView
	ExpectedState *BeaconStateView
	Eth1Timestamp Timestamp
	Eth1BlockHash Root
	Deposits      []Deposit
}

type DepositsCountMeta struct {
	DepositsCount uint64 `yaml:"deposits_count"`
}

func (c *InitializationTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.Spec = readPart.Spec()
	{
		p := readPart.Part("state.ssz")
		if p.Exists() {
			stateSize, err := p.Size()
			test_util.Check(t, err)
			state, err := AsBeaconStateView(c.Spec.BeaconState().Deserialize(codec.NewDecodingReader(p, stateSize)))
			test_util.Check(t, err)
			c.ExpectedState = state
		} else {
			// expecting a failed genesis
			c.ExpectedState = nil
		}
	}
	{
		p := readPart.Part("eth1_block_hash.yaml")
		dec := yaml.NewDecoder(p)
		var blockHash string
		test_util.Check(t, dec.Decode(&blockHash))
		test_util.Check(t, p.Close())
		_, err := hex.Decode(c.Eth1BlockHash[:], []byte(blockHash)[2:])
		test_util.Check(t, err)
	}
	{
		p := readPart.Part("eth1_timestamp.yaml")
		dec := yaml.NewDecoder(p)
		var timestamp Timestamp
		test_util.Check(t, dec.Decode(&timestamp))
		test_util.Check(t, p.Close())
		c.Eth1Timestamp = timestamp
	}
	m := &DepositsCountMeta{}
	{
		p := readPart.Part("meta.yaml")
		dec := yaml.NewDecoder(p)
		test_util.Check(t, dec.Decode(&m))
		test_util.Check(t, p.Close())
	}
	{
		for i := uint64(0); i < m.DepositsCount; i++ {
			var dep Deposit
			test_util.LoadSSZ(t, fmt.Sprintf("deposits_%d", i), &dep, readPart)
			c.Deposits = append(c.Deposits, dep)
		}
	}
}

func (c *InitializationTestCase) Run() error {
	res, _, err := c.Spec.GenesisFromEth1(c.Eth1BlockHash, c.Eth1Timestamp, c.Deposits, false)
	if err != nil {
		return err
	}
	c.GenesisState = res
	return nil
}

func (c *InitializationTestCase) ExpectingFailure() bool {
	return c.ExpectedState == nil
}

func (c *InitializationTestCase) Check(t *testing.T) {
	if c.ExpectingFailure() {
		t.Errorf("was expecting failure, but no error was raised")
	} else {
		diff, err := test_util.CompareStates(c.Spec, c.GenesisState, c.ExpectedState)
		if err != nil {
			t.Fatal(err)
		}
		if diff != "" {
			t.Errorf("genesis result does not match expectation!\n%s", diff)
		}
	}
}

func TestInitialization(t *testing.T) {
	test_util.RunTransitionTest(t, "genesis", "initialization",
		func() test_util.TransitionTest { return new(InitializationTestCase) })
}
