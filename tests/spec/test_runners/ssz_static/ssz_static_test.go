package ssz_static

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"github.com/protolambda/zrnt/eth2/beacon/attestations"
	"github.com/protolambda/zrnt/eth2/beacon/deposits"
	"github.com/protolambda/zrnt/eth2/beacon/eth1"
	"github.com/protolambda/zrnt/eth2/beacon/exits"
	"github.com/protolambda/zrnt/eth2/beacon/header"
	"github.com/protolambda/zrnt/eth2/beacon/history"
	"github.com/protolambda/zrnt/eth2/beacon/slashings/attslash"
	"github.com/protolambda/zrnt/eth2/beacon/slashings/propslash"
	"github.com/protolambda/zrnt/eth2/beacon/validator"
	"github.com/protolambda/zrnt/eth2/beacon/versioning"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"github.com/protolambda/zssz"
	"github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/types"
	"gopkg.in/yaml.v2"
	"testing"
)

type SSZStaticTestCase struct {
	TypeName string

	SSZ        types.SSZ
	Value      interface{}
	Serialized []byte

	Root        Root
	SigningRoot Root
}

func (testCase *SSZStaticTestCase) Run(t *testing.T) {
	// deserialization is the pre-requisite
	{
		r := bytes.NewReader(testCase.Serialized)
		if err := zssz.Decode(r, uint64(len(testCase.Serialized)), testCase.Value, testCase.SSZ); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("serialization", func(t *testing.T) {
		var data []byte
		{
			var buf bytes.Buffer
			bufWriter := bufio.NewWriter(&buf)
			if _, err := zssz.Encode(bufWriter, testCase.Value, testCase.SSZ); err != nil {
				t.Error(err)
				return
			}
			if err := bufWriter.Flush(); err != nil {
				t.Error(err)
				return
			}
			data = buf.Bytes()
		}

		if len(data) != len(testCase.Serialized) {
			t.Errorf("encoded data has different length: %d (spec) <-> %d (zrnt)\nspec: %x\nzrnt: %x", len(testCase.Serialized), len(data), testCase.Serialized, data)
			return
		}
		for i := 0; i < len(data); i++ {
			if data[i] != testCase.Serialized[i] {
				t.Errorf("byte i: %d differs: %d (spec) <-> %d (zrnt)\nspec: %x\nzrnt: %x", i, testCase.Serialized[i], data[i], testCase.Serialized, data)
				return
			}
		}
	})

	t.Run("hash_tree_root", func(t *testing.T) {
		hfn := hashing.GetHashFn()
		root := Root(zssz.HashTreeRoot(htr.HashFn(hfn), testCase.Value, testCase.SSZ))
		if root != testCase.Root {
			t.Errorf("hash-tree-roots differ: %x (spec) <-> %x (zrnt)", testCase.Root, root)
			return
		}
	})

	if testCase.SigningRoot != (Root{}) {
		signedSSZ, ok := testCase.SSZ.(types.SignedSSZ)
		if ok {
			t.Run("signing_root", func(t *testing.T) {
				hfn := hashing.GetHashFn()
				root := Root(zssz.SigningRoot(htr.HashFn(hfn), testCase.Value, signedSSZ))
				if root != testCase.SigningRoot {
					t.Errorf("signing-roots differ: %x (spec) <-> %x (zrnt)", testCase.SigningRoot, root)
				}
			})
		}
	}
}

type ObjAllocator func() interface{}

type ObjData struct {
	TypeName string
	Alloc    ObjAllocator
}

var objs = []*ObjData{
	{TypeName: "Fork", Alloc: func() interface{} { return new(versioning.Fork) }},
	{TypeName: "Eth1Data", Alloc: func() interface{} { return new(eth1.Eth1Data) }},
	{TypeName: "AttestationData", Alloc: func() interface{} { return new(attestations.AttestationData) }},
	{TypeName: "IndexedAttestation", Alloc: func() interface{} { return new(attestations.IndexedAttestation) }},
	{TypeName: "DepositData", Alloc: func() interface{} { return new(deposits.DepositData) }},
	{TypeName: "BeaconBlockHeader", Alloc: func() interface{} { return new(header.BeaconBlockHeader) }},
	{TypeName: "Validator", Alloc: func() interface{} { return new(validator.Validator) }},
	{TypeName: "PendingAttestation", Alloc: func() interface{} { return new(attestations.PendingAttestation) }},
	{TypeName: "HistoricalBatch", Alloc: func() interface{} { return new(history.HistoricalBatch) }},
	{TypeName: "ProposerSlashing", Alloc: func() interface{} { return new(propslash.ProposerSlashing) }},
	{TypeName: "AttesterSlashing", Alloc: func() interface{} { return new(attslash.AttesterSlashing) }},
	{TypeName: "Attestation", Alloc: func() interface{} { return new(attestations.Attestation) }},
	{TypeName: "Deposit", Alloc: func() interface{} { return new(deposits.Deposit) }},
	{TypeName: "VoluntaryExit", Alloc: func() interface{} { return new(exits.VoluntaryExit) }},
	{TypeName: "BeaconBlockBody", Alloc: func() interface{} { return new(phase0.BeaconBlockBody) }},
	{TypeName: "BeaconBlock", Alloc: func() interface{} { return new(phase0.BeaconBlock) }},
	{TypeName: "BeaconState", Alloc: func() interface{} { return new(phase0.BeaconState) }},
}

type RootsYAML struct {
	Root        string `yaml:"root"`
	SigningRoot string `yaml:"signing_root"`
}

func (obj *ObjData) RunHandler(t *testing.T) {
	test_util.RunHandler(t, "ssz_static/"+obj.TypeName, func(t *testing.T, readPart test_util.TestPartReader) {
		c := &SSZStaticTestCase{
			TypeName: obj.TypeName,
		}

		// Allocate an empty value to decode into later for testing.
		c.Value = obj.Alloc()

		// Get the SSZ type
		c.SSZ = zssz.GetSSZ(c.Value)

		// Load the SSZ encoded data as a bytes array. The test will serialize it both ways.
		{
			p := readPart("serialized.ssz")
			size, err := p.Size()
			test_util.Check(t, err)
			buf := new(bytes.Buffer)
			n, err := buf.ReadFrom(p)
			test_util.Check(t, err)
			test_util.Check(t, p.Close())
			if uint64(n) != size {
				t.Errorf("could not read full serialized data")
			}
			c.Serialized = buf.Bytes()
		}

		{
			p := readPart("roots.yaml")
			dec := yaml.NewDecoder(p)
			roots := &RootsYAML{}
			test_util.Check(t, dec.Decode(roots))
			test_util.Check(t, p.Close())
			{
				root, err := hex.DecodeString(roots.Root[2:])
				test_util.Check(t, err)
				copy(c.Root[:], root)
			}
			if len(roots.SigningRoot) >= 2 {
				root, err := hex.DecodeString(roots.SigningRoot[2:])
				test_util.Check(t, err)
				copy(c.SigningRoot[:], root)
			}
		}

		// Run the test case
		c.Run(t)

	}, PRESET_NAME)
}

func TestSSZStatic(t *testing.T) {
	t.Parallel()
	for _, o := range objs {
		o.RunHandler(t)
	}
}
