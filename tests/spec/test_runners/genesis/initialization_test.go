package sanity

import (
	"encoding/hex"
	"fmt"
	"github.com/protolambda/messagediff"
	"github.com/protolambda/zrnt/eth2/beacon/deposits"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"gopkg.in/yaml.v2"
	"testing"
)

type InitializationTestCase struct {
	GenesisState  *phase0.BeaconState
	ExpectedState *phase0.BeaconState
	Eth1Timestamp Timestamp
	Eth1BlockHash Root
	Deposits      []deposits.Deposit
}

type DepositsCountMeta struct {
	DepositsCount uint64 `yaml:"deposits_count"`
}

func (c *InitializationTestCase) Load(t *testing.T, readPart test_util.TestPartReader) {
	c.ExpectedState = new(phase0.BeaconState)
	if !test_util.LoadSSZ(t, "state", c.ExpectedState, phase0.BeaconStateSSZ, readPart) {
		// expecting a failed genesis
		c.ExpectedState = nil
	}
	{
		p := readPart("eth1_block_hash.yaml")
		dec := yaml.NewDecoder(p)
		var blockHash string
		test_util.Check(t, dec.Decode(&blockHash))
		test_util.Check(t, p.Close())
		_, err := hex.Decode(c.Eth1BlockHash[:], []byte(blockHash)[2:])
		test_util.Check(t, err)
	}
	{
		p := readPart("eth1_timestamp.yaml")
		dec := yaml.NewDecoder(p)
		var timestamp Timestamp
		test_util.Check(t, dec.Decode(&timestamp))
		test_util.Check(t, p.Close())
		c.Eth1Timestamp = timestamp
	}
	m := &DepositsCountMeta{}
	{
		p := readPart("meta.yaml")
		dec := yaml.NewDecoder(p)
		test_util.Check(t, dec.Decode(&m))
		test_util.Check(t, p.Close())
	}
	{
		for i := uint64(0); i < m.DepositsCount; i++ {
			var dep deposits.Deposit
			test_util.LoadSSZ(t, fmt.Sprintf("deposits_%d", i), &dep, deposits.DepositSSZ, readPart)
			c.Deposits = append(c.Deposits, dep)
		}
	}
}

func (c *InitializationTestCase) Run() error {
	res, err := phase0.GenesisFromEth1(c.Eth1BlockHash, c.Eth1Timestamp, c.Deposits, true)
	if err != nil {
		return err
	}
	c.GenesisState = res.BeaconState
	return nil
}

func (c *InitializationTestCase) ExpectingFailure() bool {
	return c.ExpectedState == nil
}

func (c *InitializationTestCase) Check(t *testing.T) {
	if c.ExpectingFailure() {
		t.Errorf("was expecting failure, but no error was raised")
	} else if diff, equal := messagediff.PrettyDiff(c.GenesisState, c.ExpectedState, messagediff.SliceWeakEmptyOption{}); !equal {
		t.Errorf("genesis result does not match expectation!\n%s", diff)
	}
}

func TestInitialization(t *testing.T) {
	test_util.RunTransitionTest(t, "genesis", "initialization",
		func() test_util.TransitionTest { return new(InitializationTestCase) })
}
