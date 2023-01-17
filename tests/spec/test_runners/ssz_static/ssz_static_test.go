package ssz_static

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/protolambda/zrnt/eth2/beacon/capella"

	"github.com/golang/snappy"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"gopkg.in/yaml.v3"

	"github.com/protolambda/zrnt/eth2/beacon/altair"
	"github.com/protolambda/zrnt/eth2/beacon/bellatrix"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/zrnt/tests/spec/test_util"
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

var objs = map[test_util.ForkName]map[string]ObjAllocator{
	"phase0":    {},
	"altair":    {},
	"bellatrix": {},
	"capella":   {},
}

func init() {
	base := map[string]ObjAllocator{
		"AggregateAndProof": func() interface{} { return new(phase0.AggregateAndProof) },
		"Attestation":       func() interface{} { return new(phase0.Attestation) },
		"AttestationData":   func() interface{} { return new(phase0.AttestationData) },
		"AttesterSlashing":  func() interface{} { return new(phase0.AttesterSlashing) },
		"BeaconBlockHeader": func() interface{} { return new(common.BeaconBlockHeader) },
		"Checkpoint":        func() interface{} { return new(common.Checkpoint) },
		"Deposit":           func() interface{} { return new(common.Deposit) },
		"DepositData":       func() interface{} { return new(common.DepositData) },
		//"Eth1Block": func() interface{} { return new(common.Eth1Block) }, // phase0 validator spec remnant
		"Eth1Data":                func() interface{} { return new(common.Eth1Data) },
		"Fork":                    func() interface{} { return new(common.Fork) },
		"ForkData":                func() interface{} { return new(common.ForkData) },
		"HistoricalBatch":         func() interface{} { return new(phase0.HistoricalBatch) },
		"IndexedAttestation":      func() interface{} { return new(phase0.IndexedAttestation) },
		"PendingAttestation":      func() interface{} { return new(phase0.PendingAttestation) },
		"ProposerSlashing":        func() interface{} { return new(phase0.ProposerSlashing) },
		"SignedAggregateAndProof": func() interface{} { return new(phase0.SignedAggregateAndProof) },
		"SignedBeaconBlockHeader": func() interface{} { return new(common.SignedBeaconBlockHeader) },
		"SignedVoluntaryExit":     func() interface{} { return new(phase0.SignedVoluntaryExit) },
		//"SigningData": func() interface{} { return new(common.SigningData) },  // not really encoded/decoded, just HTR
		"Validator":     func() interface{} { return new(phase0.Validator) },
		"VoluntaryExit": func() interface{} { return new(phase0.VoluntaryExit) },
	}
	for k, v := range base {
		objs["phase0"][k] = v
		objs["altair"][k] = v
		objs["bellatrix"][k] = v
		objs["capella"][k] = v
	}
	objs["phase0"]["BeaconBlockBody"] = func() interface{} { return new(phase0.BeaconBlockBody) }
	objs["phase0"]["BeaconBlock"] = func() interface{} { return new(phase0.BeaconBlock) }
	objs["phase0"]["BeaconState"] = func() interface{} { return new(phase0.BeaconState) }
	objs["phase0"]["SignedBeaconBlock"] = func() interface{} { return new(phase0.SignedBeaconBlock) }

	objs["altair"]["BeaconBlockBody"] = func() interface{} { return new(altair.BeaconBlockBody) }
	objs["altair"]["BeaconBlock"] = func() interface{} { return new(altair.BeaconBlock) }
	objs["altair"]["BeaconState"] = func() interface{} { return new(altair.BeaconState) }
	objs["altair"]["SignedBeaconBlock"] = func() interface{} { return new(altair.SignedBeaconBlock) }
	objs["altair"]["SyncAggregate"] = func() interface{} { return new(altair.SyncAggregate) }

	objs["altair"]["LightClientSnapshot"] = func() interface{} { return new(altair.LightClientSnapshot) }
	objs["altair"]["LightClientUpdate"] = func() interface{} { return new(altair.LightClientUpdate) }
	objs["altair"]["SyncAggregatorSelectionData"] = func() interface{} { return new(altair.SyncAggregatorSelectionData) }
	objs["altair"]["SyncCommitteeContribution"] = func() interface{} { return new(altair.SyncCommitteeContribution) }
	objs["altair"]["ContributionAndProof"] = func() interface{} { return new(altair.ContributionAndProof) }
	objs["altair"]["SignedContributionAndProof"] = func() interface{} { return new(altair.SignedContributionAndProof) }
	objs["altair"]["SyncCommitteeMessage"] = func() interface{} { return new(altair.SyncCommitteeMessage) }
	objs["altair"]["SyncCommittee"] = func() interface{} { return new(common.SyncCommittee) }

	objs["bellatrix"]["BeaconBlockBody"] = func() interface{} { return new(bellatrix.BeaconBlockBody) }
	objs["bellatrix"]["BeaconBlock"] = func() interface{} { return new(bellatrix.BeaconBlock) }
	objs["bellatrix"]["BeaconState"] = func() interface{} { return new(bellatrix.BeaconState) }
	objs["bellatrix"]["SignedBeaconBlock"] = func() interface{} { return new(bellatrix.SignedBeaconBlock) }
	objs["bellatrix"]["ExecutionPayload"] = func() interface{} { return new(bellatrix.ExecutionPayload) }
	objs["bellatrix"]["ExecutionPayloadHeader"] = func() interface{} { return new(bellatrix.ExecutionPayloadHeader) }
	//objs["bellatrix"]["PowBlock"] = func() interface{} { return new(bellatrix.PowBlock) }

	objs["capella"]["BeaconBlockBody"] = func() interface{} { return new(capella.BeaconBlockBody) }
	objs["capella"]["BeaconBlock"] = func() interface{} { return new(capella.BeaconBlock) }
	objs["capella"]["BeaconState"] = func() interface{} { return new(capella.BeaconState) }
	objs["capella"]["SignedBeaconBlock"] = func() interface{} { return new(capella.SignedBeaconBlock) }
	objs["capella"]["ExecutionPayload"] = func() interface{} { return new(capella.ExecutionPayload) }
	objs["capella"]["ExecutionPayloadHeader"] = func() interface{} { return new(capella.ExecutionPayloadHeader) }
	objs["capella"]["Withdrawal"] = func() interface{} { return new(common.Withdrawal) }
	objs["capella"]["BLSToExecutionChange"] = func() interface{} { return new(common.BLSToExecutionChange) }
	objs["capella"]["SignedBLSToExecutionChange"] = func() interface{} { return new(common.SignedBLSToExecutionChange) }
}

type RootsYAML struct {
	Root string `yaml:"root"`
}

func runSSZStaticTest(fork test_util.ForkName, name string, alloc ObjAllocator, spec *common.Spec) func(t *testing.T) {
	return func(t *testing.T) {
		test_util.RunHandler(t, "ssz_static/"+name, func(t *testing.T, forkName test_util.ForkName, readPart test_util.TestPartReader) {
			c := &SSZStaticTestCase{
				Spec:     readPart.Spec(),
				TypeName: name,
			}

			// Allocate an empty value to decode into later for testing.
			c.Value = alloc()

			// Load the SSZ encoded data as a bytes array. The test will serialize it both ways.
			{
				p := readPart.Part("serialized.ssz_snappy")
				data, err := ioutil.ReadAll(p)
				test_util.Check(t, err)
				uncompressed, err := snappy.Decode(nil, data)
				test_util.Check(t, err)
				test_util.Check(t, p.Close())
				test_util.Check(t, err)
				c.Serialized = uncompressed
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

		}, spec, fork)
	}
}

func TestSSZStatic(t *testing.T) {
	t.Parallel()
	t.Run("minimal", func(t *testing.T) {
		for fork, objByName := range objs {
			t.Run(string(fork), func(t *testing.T) {
				for k, v := range objByName {
					t.Run(k, runSSZStaticTest(fork, k, v, configs.Minimal))
				}
			})
		}
	})
	t.Run("mainnet", func(t *testing.T) {
		for fork, objByName := range objs {
			t.Run(string(fork), func(t *testing.T) {
				for k, v := range objByName {
					t.Run(k, runSSZStaticTest(fork, k, v, configs.Mainnet))
				}
			})
		}
	})
}
