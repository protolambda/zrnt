package ssz

import (
	"fmt"
	"github.com/protolambda/go-beacon-transition/eth2"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"testing"
)

func TestHash_tree_root(t *testing.T) {
	es := beacon.BeaconState{}
	es.Latest_randao_mixes = make([]eth2.Bytes32, eth2.LATEST_RANDAO_MIXES_LENGTH)
	blockHash := Hash_tree_root(es)
	fmt.Printf("%x\n", blockHash)
}