package main

import (
	"encoding/binary"
	"fmt"
	"github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/experimental/tree"
	. "github.com/protolambda/zrnt/experimental/views"
)

// Experimental code! Everything a tree and cached by default.
//  - anything can be a node
//  - commits are wrapped with navigation to fetch getters/setters
//  - commits can be made read-only
//  - modifications in a subtree can be batched: do not rebind each all the way up to the root of the tree
//  - views can be injected in-between commits
//     - views can be typed
//     - views are not copied, they proxy rebinds up, and track the latest binding.
//        - views can be used as ways to pass the tree as regular mutable object
//  - Vector/List/Container views supported. Basic types do not need a view: they are immutable and a node by default (just define a ComputeRoot() to convert it into 32 bytes).
//  - views can be overlayed on existing trees
//    - overlay on incomplete tree == partial
//  - Views to be implemented still:
//     - Bitvector
//     - Bitlist
//     - Union
//     - Basic-lists

type Slot uint64

func (s Slot) ComputeRoot(h HashFn) (out Root) {
	binary.LittleEndian.PutUint64(out[:], uint64(s))
	return
}

type Block struct {
	*ContainerView
	// Not included in default hash-tree-root
	Signature core.BLSSignature
}

func NewBlock() (b *Block) {
	return &Block{ContainerView: Container(
		Slot(0),
		&Root{},
		&Root{},
		NewBlockBody(),
	)}
}
func (b *Block) Slot() Slot { return b.Get(0).(Slot) }

type BlockBody struct {
	*ContainerView
}

func NewBlockBody() (b *BlockBody) {
	return &BlockBody{Container(
		Slot(0),
		&Root{},
		// TODO operations lists
	)}
}

func main() {
	b := NewBlock()
	b.Set(0)(&Root{1})
	fmt.Printf("%x\n", b.ComputeRoot(Hash))
	b.Set(0)(&Root{2})
	fmt.Printf("%x\n", b.ComputeRoot(Hash))
	b.Set(0)(&Root{1})
	fmt.Printf("%x\n", b.ComputeRoot(Hash))
}
