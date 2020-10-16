package api

import (
	"encoding/json"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/configs"
	"github.com/protolambda/ztyp/tree"
	"io/ioutil"
	"testing"
)

func TestStateLoad(t *testing.T) {
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
}

func TestSignedBeaconBlockLoad(t *testing.T) {
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
}

func TestSignedBeaconHeaderLoad(t *testing.T) {
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
}
