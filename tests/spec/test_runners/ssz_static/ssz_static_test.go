package ssz_static

import (
	"bytes"
	"encoding/hex"
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"gopkg.in/yaml.v3"
	"testing"
)

type SSZStaticTestCase struct {
	TypeName   string
	Spec       *Spec
	Value      interface{}
	Serialized []byte

	Root Root
}

func (testCase *SSZStaticTestCase) Run(t *testing.T) {
	// deserialization is the pre-requisite
	{
		r := bytes.NewReader(testCase.Serialized)
		if obj, ok := testCase.Value.(SpecObj); ok {
			if err := obj.Deserialize(testCase.Spec, codec.NewDecodingReader(r, uint64(len(testCase.Serialized)))); err != nil {
				t.Fatal(err)
			}
		} else if des, ok := testCase.Value.(codec.Deserializable); ok {
			if err := des.Deserialize(codec.NewDecodingReader(r, uint64(len(testCase.Serialized)))); err != nil {
				t.Fatal(err)
			}
		} else {
			t.Fatalf("type %s cannot be deserialized", testCase.TypeName)
		}
	}

	t.Run("serialization", func(t *testing.T) {
		var data []byte
		{
			var buf bytes.Buffer
			if obj, ok := testCase.Value.(SpecObj); ok {
				if err := obj.Serialize(testCase.Spec, codec.NewEncodingWriter(&buf)); err != nil {
					t.Fatal(err)
				}
			} else if ser, ok := testCase.Value.(codec.Serializable); ok {
				if err := ser.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Fatalf("type %s cannot be serialized", testCase.TypeName)
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
		hfn := tree.GetHashFn()

		var root Root
		if obj, ok := testCase.Value.(SpecObj); ok {
			root = obj.HashTreeRoot(testCase.Spec, hfn)
		} else if v, ok := testCase.Value.(tree.HTR); ok {
			root = v.HashTreeRoot(hfn)
		} else {
			t.Fatalf("type %s cannot be serialized", testCase.TypeName)
		}
		if root != testCase.Root {
			t.Errorf("hash-tree-roots differ: %x (spec) <-> %x (zrnt)", testCase.Root, root)
			return
		}
	})
}

type ObjAllocator func() interface{}

type ObjData struct {
	TypeName string
	Alloc    ObjAllocator
}

var objs = []*ObjData{
	{TypeName: "Fork", Alloc: func() interface{} { return new(Fork) }},
	{TypeName: "Eth1Data", Alloc: func() interface{} { return new(Eth1Data) }},
	{TypeName: "AttestationData", Alloc: func() interface{} { return new(AttestationData) }},
	{TypeName: "IndexedAttestation", Alloc: func() interface{} { return new(IndexedAttestation) }},
	{TypeName: "DepositData", Alloc: func() interface{} { return new(DepositData) }},
	{TypeName: "BeaconBlockHeader", Alloc: func() interface{} { return new(BeaconBlockHeader) }},
	{TypeName: "Validator", Alloc: func() interface{} { return new(Validator) }},
	{TypeName: "PendingAttestation", Alloc: func() interface{} { return new(PendingAttestation) }},
	{TypeName: "HistoricalBatch", Alloc: func() interface{} { return new(HistoricalBatch) }},
	{TypeName: "ProposerSlashing", Alloc: func() interface{} { return new(ProposerSlashing) }},
	{TypeName: "AttesterSlashing", Alloc: func() interface{} { return new(AttesterSlashing) }},
	{TypeName: "Attestation", Alloc: func() interface{} { return new(Attestation) }},
	{TypeName: "Deposit", Alloc: func() interface{} { return new(Deposit) }},
	{TypeName: "VoluntaryExit", Alloc: func() interface{} { return new(VoluntaryExit) }},
	{TypeName: "BeaconBlockBody", Alloc: func() interface{} { return new(BeaconBlockBody) }},
	{TypeName: "BeaconBlock", Alloc: func() interface{} { return new(BeaconBlock) }},
	{TypeName: "BeaconState", Alloc: func() interface{} { return new(BeaconState) }},
}

type RootsYAML struct {
	Root string `yaml:"root"`
}

func (obj *ObjData) runSSZStaticTest(spec *Spec) func(t *testing.T) {
	return func(t *testing.T) {

		test_util.RunHandler(t, "ssz_static/"+obj.TypeName, func(t *testing.T, readPart test_util.TestPartReader) {
			c := &SSZStaticTestCase{
				TypeName: obj.TypeName,
			}

			// Allocate an empty value to decode into later for testing.
			c.Value = obj.Alloc()

			// Load the SSZ encoded data as a bytes array. The test will serialize it both ways.
			{
				p := readPart.Part("serialized.ssz")
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
				p := readPart.Part("roots.yaml")
				dec := yaml.NewDecoder(p)
				roots := &RootsYAML{}
				test_util.Check(t, dec.Decode(roots))
				test_util.Check(t, p.Close())
				{
					root, err := hex.DecodeString(roots.Root[2:])
					test_util.Check(t, err)
					copy(c.Root[:], root)
				}
			}

			// Run the test case
			c.Run(t)

		}, spec)
	}
}

func (obj *ObjData) RunHandler(t *testing.T) {
	t.Run("minimal", obj.runSSZStaticTest(configs.Minimal))
	t.Run("mainnet", obj.runSSZStaticTest(configs.Mainnet))
}

func TestSSZStatic(t *testing.T) {
	t.Parallel()
	for _, o := range objs {
		o.RunHandler(t)
	}
}
