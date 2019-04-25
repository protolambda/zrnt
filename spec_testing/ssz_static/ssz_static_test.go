package ssz_static

import (
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zrnt/spec_testing"
	"testing"
)

type SSZStaticTestCase struct {
	TypeName   string
	Value      interface{}
	Serialized Bytes
	Root       Root
	SigningRoot Root
}

func (testCase *SSZStaticTestCase) Run(t *testing.T) {
	t.Run(testCase.TypeName, func(t *testing.T) {
		t.Run("serialization", func(t *testing.T) {
			encoded := ssz.SSZEncode(testCase.Value)
			if len(encoded) != len(testCase.Serialized) {
				encodedBytes := Bytes(encoded)
				t.Fatalf("encoded data has different length: %d (spec) <-> %d (zrnt)\nspec: %s\nzrnt: %s", len(testCase.Serialized), len(encoded), testCase.Serialized.String(), encodedBytes.String())
			}
			for i := 0; i < len(encoded); i++ {
				if encoded[i] != testCase.Serialized[i] {
					encodedBytes := Bytes(encoded)
					t.Fatalf("byte i: %d differs: %d (spec) <-> %d (zrnt)\nspec: %s\nzrnt: %s", i, testCase.Serialized[i], encoded[i], testCase.Serialized.String(), encodedBytes.String())
				}
			}
		})
		t.Run("hash_tree_root", func(t *testing.T) {
			root := ssz.HashTreeRoot(testCase.Value)
			if root != testCase.Root {
				t.Fatalf("hash-tree-roots differ: %s (spec) <-> %s (zrnt)", testCase.Root.String(), root.String())
			}
		})
		if testCase.SigningRoot != (Root{}) {
			t.Run("signing_root", func(t *testing.T) {
				root := ssz.SigningRoot(testCase.Value)
				if root != testCase.SigningRoot {
					t.Fatalf("signing-roots differ: %s (spec) <-> %s (zrnt)", testCase.SigningRoot.String(), root.String())
				}
			})
		}
	})
}

type ObjAllocator func() interface{}

var allocators = map[string]ObjAllocator{
	"Fork": func() interface{} { return new(Fork)},
	"Crosslink": func() interface{} { return new(Crosslink)},
	"Eth1Data": func() interface{} { return new(Eth1Data)},
	"AttestationData": func() interface{} { return new(AttestationData)},
	"AttestationDataAndCustodyBit": func() interface{} { return new(AttestationDataAndCustodyBit)},
	"IndexedAttestation": func() interface{} { return new(IndexedAttestation)},
	"DepositData": func() interface{} { return new(DepositData)},
	"BeaconBlockHeader": func() interface{} { return new(BeaconBlockHeader)},
	"Validator": func() interface{} { return new(Validator)},
	"PendingAttestation": func() interface{} { return new(PendingAttestation)},
	"HistoricalBatch": func() interface{} { return new(HistoricalBatch)},
	"ProposerSlashing": func() interface{} { return new(ProposerSlashing)},
	"AttesterSlashing": func() interface{} { return new(AttesterSlashing)},
	"Attestation": func() interface{} { return new(Attestation)},
	"Deposit": func() interface{} { return new(Deposit)},
	"VoluntaryExit": func() interface{} { return new(VoluntaryExit)},
	"Transfer": func() interface{} { return new(Transfer)},
	"BeaconBlockBody": func() interface{} { return new(BeaconBlockBody)},
	"BeaconBlock": func() interface{} { return new(BeaconBlock)},
	"BeaconState": func() interface{} { return new(BeaconState)},
}

func TestSSZStatic(t *testing.T) {
	spec_testing.RunSuitesInPath("../../../eth2.0-specs/eth2.0-spec-tests/tests/ssz_static/core/",
		func(raw interface{}) interface{} {
			data := raw.(map[string]interface{})
			name := data["TypeName"].(string)
			valueAllocator := allocators[name]
			testCase := new(SSZStaticTestCase)
			testCase.Value = valueAllocator()
			return testCase
		}, t)
}
