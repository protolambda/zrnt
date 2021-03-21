package ssz_static

import (
	"bytes"
	"encoding/hex"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"gopkg.in/yaml.v3"
	"testing"
)

type SSZStaticTestCase struct {
	TypeName   string
	Spec       *common.Spec
	Value      interface{}
	Serialized []byte

	Root common.Root
}

func (testCase *SSZStaticTestCase) Run(t *testing.T) {
	// deserialization is the pre-requisite
	{
		r := bytes.NewReader(testCase.Serialized)
		if obj, ok := testCase.Value.(common.SpecObj); ok {
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
			if obj, ok := testCase.Value.(common.SpecObj); ok {
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

		var root common.Root
		if obj, ok := testCase.Value.(common.SpecObj); ok {
			root = obj.HashTreeRoot(testCase.Spec, hfn)
		} else if v, ok := testCase.Value.(tree.HTR); ok {
			root = v.HashTreeRoot(hfn)
		} else {
			t.Fatalf("type %s cannot be serialized", testCase.TypeName)
		}
		if root != testCase.Root {
			t.Errorf("hash-tree-roots differ: %s (spec) <-> %s (zrnt)", testCase.Root, root)
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
	{TypeName: "Fork", Alloc: func() interface{} { return new(common.Fork) }},
	{TypeName: "Eth1Data", Alloc: func() interface{} { return new(common.Eth1Data) }},
	{TypeName: "AttestationData", Alloc: func() interface{} { return new(phase0.AttestationData) }},
	{TypeName: "IndexedAttestation", Alloc: func() interface{} { return new(phase0.IndexedAttestation) }},
	{TypeName: "DepositData", Alloc: func() interface{} { return new(common.DepositData) }},
	{TypeName: "BeaconBlockHeader", Alloc: func() interface{} { return new(common.BeaconBlockHeader) }},
	{TypeName: "Validator", Alloc: func() interface{} { return new(phase0.Validator) }},
	{TypeName: "PendingAttestation", Alloc: func() interface{} { return new(phase0.PendingAttestation) }},
	{TypeName: "HistoricalBatch", Alloc: func() interface{} { return new(phase0.HistoricalBatch) }},
	{TypeName: "ProposerSlashing", Alloc: func() interface{} { return new(phase0.ProposerSlashing) }},
	{TypeName: "AttesterSlashing", Alloc: func() interface{} { return new(phase0.AttesterSlashing) }},
	{TypeName: "Attestation", Alloc: func() interface{} { return new(phase0.Attestation) }},
	{TypeName: "Deposit", Alloc: func() interface{} { return new(common.Deposit) }},
	{TypeName: "VoluntaryExit", Alloc: func() interface{} { return new(phase0.VoluntaryExit) }},
	{TypeName: "BeaconBlockBody", Alloc: func() interface{} { return new(phase0.BeaconBlockBody) }},
	{TypeName: "BeaconBlock", Alloc: func() interface{} { return new(phase0.BeaconBlock) }},
	{TypeName: "BeaconState", Alloc: func() interface{} { return new(phase0.BeaconState) }},
}

type RootsYAML struct {
	Root string `yaml:"root"`
}

func (obj *ObjData) runSSZStaticTest(spec *common.Spec) func(t *testing.T) {
	return func(t *testing.T) {

		test_util.RunHandler(t, "ssz_static/"+obj.TypeName, func(t *testing.T, readPart test_util.TestPartReader) {
			c := &SSZStaticTestCase{
				Spec:     readPart.Spec(),
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
