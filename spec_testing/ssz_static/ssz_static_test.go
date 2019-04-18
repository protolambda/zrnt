package operations

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/data_types"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zrnt/spec_testing"
	"testing"
)

type SSZStaticTestCase struct {
	TypeName   string
	Value      interface{}
	Serialized data_types.Bytes
	Root       data_types.Root
}

func (testCase *SSZStaticTestCase) Run(t *testing.T) {
	t.Run("serialization", func(t *testing.T) {
		encoded := ssz.SSZEncode(testCase.Value)
		if len(encoded) != len(testCase.Serialized) {
			t.Fatalf("encoded data has different length: %d (spec) <-> %d (zrnt)", len(testCase.Serialized), len(encoded))
		}
		for i := 0; i < len(encoded); i++ {
			if encoded[i] != testCase.Serialized[i] {
				t.Fatalf("byte i: %d differs: %d (spec) <-> %d (zrnt)", i, testCase.Serialized[i], encoded[i])
			}
		}
	})
	t.Run("hash_tree_root", func(t *testing.T) {
		root := data_types.Root(ssz.HashTreeRoot(testCase.Value))
		if root != testCase.Root {
			t.Fatalf("hash-tree-roots differ: %s (spec) <-> %x (zrnt)", testCase.Root.String(), root.String())
		}
	})
}

type ObjAllocator func() interface{}

var allocators = map[string]ObjAllocator{
	"Fork": func() interface{} { return new(beacon.Fork)},
	"Crosslink": func() interface{} { return new(beacon.Crosslink)},
	"Eth1Data": func() interface{} { return new(beacon.Eth1Data)},
	"AttestationData": func() interface{} { return new(beacon.AttestationData)},
	"AttestationDataAndCustodyBit": func() interface{} { return new(beacon.AttestationDataAndCustodyBit)},
	"IndexedAttestation": func() interface{} { return new(beacon.IndexedAttestation)},
	"DepositData": func() interface{} { return new(beacon.DepositData)},
	"BeaconBlockHeader": func() interface{} { return new(beacon.BeaconBlockHeader)},
	"Validator": func() interface{} { return new(beacon.Validator)},
	"PendingAttestation": func() interface{} { return new(beacon.PendingAttestation)},
	"HistoricalBatch": func() interface{} { return new(beacon.HistoricalBatch)},
	"ProposerSlashing": func() interface{} { return new(beacon.ProposerSlashing)},
	"AttesterSlashing": func() interface{} { return new(beacon.AttesterSlashing)},
	"Attestation": func() interface{} { return new(beacon.Attestation)},
	"Deposit": func() interface{} { return new(beacon.Deposit)},
	"VoluntaryExit": func() interface{} { return new(beacon.VoluntaryExit)},
	"Transfer": func() interface{} { return new(beacon.Transfer)},
	"BeaconBlockBody": func() interface{} { return new(beacon.BeaconBlockBody)},
	"BeaconBlock": func() interface{} { return new(beacon.BeaconBlock)},
	"BeaconState": func() interface{} { return new(beacon.BeaconState)},
}

func TestSSZStatic(t *testing.T) {
	spec_testing.RunSuitesInPath("../../../eth2.0-specs/yaml_tests/ssz_static/core/",
		func(raw interface{}) interface{} {
			data := raw.(map[string]interface{})
			name := data["TypeName"].(string)
			valueAllocator := allocators[name]
			testCase := new(SSZStaticTestCase)
			testCase.Value = valueAllocator()
			return testCase
		}, t)
}
