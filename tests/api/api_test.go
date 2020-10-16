package api

import (
	"encoding/json"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/tree"
	"io/ioutil"
	"testing"
)

func TestStateJSON(t *testing.T) {
	b, err := ioutil.ReadFile("data/zinken_genesis_state.json")
	if err != nil {
		t.Fatalf("failed to open state: %v", err)
	}
	var state beacon.BeaconState
	if err := json.Unmarshal(b, &state); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	root := state.HashTreeRoot(configs.Mainnet, tree.GetHashFn())
	if root.String() != "0x7bf61f7e24ae8d6c21d61b066af5637c5e7ef5022d1deb001094d83b468610a1" {
		t.Fatalf("bad state root, decoded data must be wrong. Got state root: %s", root)
	}
	out, err := json.Marshal(&state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var state2 beacon.BeaconState
	if err := json.Unmarshal(out, &state2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	root2 := state2.HashTreeRoot(configs.Mainnet, tree.GetHashFn())
	if root != root2 {
		t.Fatalf("round trip encode/decode failed, got root: %s", root)
	}
}

func TestSignedBeaconBlockJSON(t *testing.T) {
	b, err := ioutil.ReadFile("data/zinken_signed_block.json")
	if err != nil {
		t.Fatalf("failed to open block: %v", err)
	}
	var signedBlock beacon.SignedBeaconBlock
	if err := json.Unmarshal(b, &signedBlock); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	root := signedBlock.HashTreeRoot(configs.Mainnet, tree.GetHashFn())
	if root.String() != "0x6d95c619aedf04739150f8a8230d8a825e239555bdeb0da46bb9276946242fca" {
		s, err := json.MarshalIndent(signedBlock, "  ", "  ")
		if err != nil {
			t.Errorf("failed to marshal: %v", err)
		} else {
			t.Log(string(s))
		}
		t.Fatalf("bad block root, decoded data must be wrong. Got block root: %s", root)
	}
	out, err := json.Marshal(&signedBlock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var signedBlock2 beacon.SignedBeaconBlock
	if err := json.Unmarshal(out, &signedBlock2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	root2 := signedBlock2.HashTreeRoot(configs.Mainnet, tree.GetHashFn())
	if root != root2 {
		t.Fatalf("round trip encode/decode failed, got root: %s", root)
	}
}

func TestSignedBeaconHeaderJSON(t *testing.T) {
	b, err := ioutil.ReadFile("data/zinken_signed_header.json")
	if err != nil {
		t.Fatalf("failed to open block: %v", err)
	}
	var signedHeader beacon.SignedBeaconBlockHeader
	if err := json.Unmarshal(b, &signedHeader); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	root := signedHeader.HashTreeRoot(tree.GetHashFn())
	// root of signedHeader.Message is 0x40809a6b5f53fa88948c3372cf72fafa133004e3fd2f3fde45744b8d9ea215fd,
	// but let's also check signature here.
	if root.String() != "0x6d95c619aedf04739150f8a8230d8a825e239555bdeb0da46bb9276946242fca" {
		s, err := json.MarshalIndent(signedHeader, "  ", "  ")
		if err != nil {
			t.Errorf("failed to marshal: %v", err)
		} else {
			t.Log(string(s))
		}
		t.Log(signedHeader.Message.BodyRoot.String())
		t.Fatalf("bad block root, decoded data must be wrong. Got block root: %s", root)
	}
	out, err := json.Marshal(&signedHeader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var signedHeader2 beacon.SignedBeaconBlockHeader
	if err := json.Unmarshal(out, &signedHeader2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	root2 := signedHeader2.HashTreeRoot(tree.GetHashFn())
	if root != root2 {
		t.Fatalf("round trip encode/decode failed, got root: %s", root)
	}
}
