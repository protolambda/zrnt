package go_beacon_transition

import (
	"fmt"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/transition"
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
	"testing"
)

func TestThing(t *testing.T) {
	//bl := beacon.BeaconBlock{}
	//bl.Slot = 1 << 32
	//blockHash := Hash_tree_root(bl)
	st := transition.GetGenesisBeaconState([]beacon.Deposit{}, 0, beacon.Eth1Data{})
	fmt.Printf("%x\n", ssz.Hash_tree_root(st))
	//bl := beacon.GetEmptyBlock()
	//fmt.Printf("%x\n", ssz.Hash_tree_root(bl))
}