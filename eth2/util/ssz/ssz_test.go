package ssz

import (
	"fmt"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"testing"
)

func TestHash_tree_root(t *testing.T) {
	blockHash := Hash_tree_root(beacon.BeaconBlock{})
	fmt.Println(blockHash)
}