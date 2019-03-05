package ssz

import (
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"testing"
)

func TestHash_tree_root(t *testing.T) {
	Hash_tree_root(beacon.BeaconBlock{})
}