package main

import (
	"encoding/binary"
	"fmt"
	"github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/experimental/tree"
	. "github.com/protolambda/zrnt/experimental/views"
)

// Experimental code! Everything a tree and cached by default.
// Also spaghetti, iterating on the idea/approach first, then refactoring/polishing later.

// Example usage:
//  - anything can be a node
//  - commits are wrapped with navigation to fetch getter/setter pairs of containers.
//  - commits can be made read-only
//  - modifications can be batched

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
