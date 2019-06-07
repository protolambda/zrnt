package ssz_static

import (
	"bufio"
	"bytes"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/tests/spec/test_util"
	"github.com/protolambda/zssz"
	"github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/types"
	"gopkg.in/d4l3k/messagediff.v1"
	"testing"
)


type SSZStaticTestCase struct {
	TypeName    string

	SSZ         types.SSZ
	EmptyValue  interface{}

	Value       interface{}
	Serialized  Bytes
	Root        Root
	SigningRoot Root
}

func (testCase *SSZStaticTestCase) Run(t *testing.T) {
	t.Run(testCase.TypeName, func(t *testing.T) {
		t.Run("encode", func(t *testing.T) {
			var buf bytes.Buffer
			bufWriter := bufio.NewWriter(&buf)
			if err := zssz.Encode(bufWriter, testCase.Value, testCase.SSZ); err != nil {
				t.Error(err)
				return
			}
			if err := bufWriter.Flush(); err != nil {
				t.Error(err)
				return
			}
			data := buf.Bytes()
			if len(data) != len(testCase.Serialized) {
				b := Bytes(data)
				t.Errorf("encoded data has different length: %d (spec) <-> %d (zrnt)\nspec: %s\nzrnt: %s", len(testCase.Serialized), len(data), testCase.Serialized.String(), b.String())
				return
			}
			for i := 0; i < len(data); i++ {
				if data[i] != testCase.Serialized[i] {
					b := Bytes(data)
					t.Errorf("byte i: %d differs: %d (spec) <-> %d (zrnt)\nspec: %s\nzrnt: %s", i, testCase.Serialized[i], data[i], testCase.Serialized.String(), b.String())
					return
				}
			}
		})
		t.Run("decode", func(t *testing.T) {
			r := bytes.NewReader(testCase.Serialized)
			if err := zssz.Decode(r, uint32(len(testCase.Serialized)), testCase.EmptyValue, testCase.SSZ); err != nil {
				t.Fatal(err)
			}
			if diff, equal := messagediff.PrettyDiff(testCase.EmptyValue, testCase.Value); !equal {
				t.Errorf("decode result does not match expectation!\n%s", diff)
				return
			}
		})
		t.Run("hash_tree_root", func(t *testing.T) {
			hfn := hashing.GetHashFn()
			root := Root(zssz.HashTreeRoot(htr.HashFn(hfn), testCase.Value, testCase.SSZ))
			if root != testCase.Root {
				t.Errorf("hash-tree-roots differ: %s (spec) <-> %s (zrnt)", testCase.Root.String(), root.String())
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
						t.Errorf("signing-roots differ: %s (spec) <-> %s (zrnt)", testCase.SigningRoot.String(), root.String())
					}
				})
			}
		}
	})
}

type ObjAllocator func() interface{}

type ObjData struct {
	Alloc ObjAllocator
	SSZ   types.SSZ
}

var objs = map[string]*ObjData{
	"Fork":                         {Alloc: func() interface{} { return new(Fork) }},
	"Crosslink":                    {Alloc: func() interface{} { return new(Crosslink) }},
	"Eth1Data":                     {Alloc: func() interface{} { return new(Eth1Data) }},
	"AttestationData":              {Alloc: func() interface{} { return new(AttestationData) }},
	"AttestationDataAndCustodyBit": {Alloc: func() interface{} { return new(AttestationDataAndCustodyBit) }},
	"IndexedAttestation":           {Alloc: func() interface{} { return new(IndexedAttestation) }},
	"DepositData":                  {Alloc: func() interface{} { return new(DepositData) }},
	"BeaconBlockHeader":            {Alloc: func() interface{} { return new(BeaconBlockHeader) }},
	"Validator":                    {Alloc: func() interface{} { return new(Validator) }},
	"PendingAttestation":           {Alloc: func() interface{} { return new(PendingAttestation) }},
	"HistoricalBatch":              {Alloc: func() interface{} { return new(HistoricalBatch) }},
	"ProposerSlashing":             {Alloc: func() interface{} { return new(ProposerSlashing) }},
	"AttesterSlashing":             {Alloc: func() interface{} { return new(AttesterSlashing) }},
	"Attestation":                  {Alloc: func() interface{} { return new(Attestation) }},
	"Deposit":                      {Alloc: func() interface{} { return new(Deposit) }},
	"VoluntaryExit":                {Alloc: func() interface{} { return new(VoluntaryExit) }},
	"Transfer":                     {Alloc: func() interface{} { return new(Transfer) }},
	"BeaconBlockBody":              {Alloc: func() interface{} { return new(BeaconBlockBody) }},
	"BeaconBlock":                  {Alloc: func() interface{} { return new(BeaconBlock) }},
	"BeaconState":                  {Alloc: func() interface{} { return new(BeaconState) }},
}

func init()  {
	for _, o := range objs {
		o.SSZ = zssz.GetSSZ(o.Alloc())
	}
}

func TestSSZStatic(t *testing.T) {
	test_util.RunSuitesInPath("ssz_static/core/",
		func(raw interface{}) (interface{}, interface{}) {
			data := raw.(map[string]interface{})
			for name, sszData := range data {
				objData, ok := objs[name]
				if !ok {
					panic(fmt.Sprintf("unknown ssz object type: %s", name))
				}
				testCase := &SSZStaticTestCase{TypeName: name, SSZ: objData.SSZ}
				testCase.Value = objData.Alloc()
				testCase.EmptyValue = objData.Alloc()
				return testCase, sszData
			}
			return nil, nil
		}, t)
}
