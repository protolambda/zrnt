package main

import (
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"testing"
)

func TestHeaderHash(t *testing.T) {
	block := beacon.GetEmptyBlock()
	block.StateRoot = beacon.Root{1, 2, 3}
	fmt.Printf("%x\n", ssz.HashTreeRoot(block.GetTemporaryBlockHeader()))
	fmt.Printf("%x\n", ssz.HashTreeRoot(block))
	h := block.GetTemporaryBlockHeader()
	h.StateRoot = block.StateRoot
	fmt.Printf("%x\n", ssz.HashTreeRoot(h))
	block.StateRoot = beacon.Root{}
	fmt.Printf("%x\n", ssz.HashTreeRoot(block))
}
