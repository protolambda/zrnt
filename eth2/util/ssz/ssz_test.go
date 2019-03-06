package ssz

import (
	"fmt"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"testing"
)

func TestHash_tree_root(t *testing.T) {
	bl := beacon.BeaconBlock{}
	bl.Slot = 1 << 32
	blockHash := Hash_tree_root(bl)
	fmt.Printf("%x\n", blockHash)
}