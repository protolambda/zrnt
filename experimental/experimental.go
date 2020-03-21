package main

import (
	"fmt"
	"github.com/protolambda/zrnt/eth2/phase0"
	"github.com/protolambda/ztyp/tree"
	"github.com/protolambda/ztyp/view"
)

func printNode(node tree.Node, path string) {
	//if root, ok := node.(*tree.Root); ok {
	//	//fmt.Printf("\"%x_%s\"\n", *root, path)
	//	return
	//}
	if commit, ok := node.(*tree.Commit); ok {
		left := path + "0"
		right := path + "1"
		root := commit.MerkleRoot(tree.Hash)
		leftRoot := commit.Left.MerkleRoot(tree.Hash)
		rightRoot := commit.Right.MerkleRoot(tree.Hash)
		fmt.Printf("\"%x_%s\", \"%x_%s\"\n", root, path, leftRoot, left)
		fmt.Printf("\"%x_%s\", \"%x_%s\"\n", root, path, rightRoot, right)
		printNode(commit.Left, left)
		printNode(commit.Right, right)
	}
}

func main() {
	state := phase0.BeaconStateType.New(func(v view.View) error {
		fmt.Println("new state change: ", v)
		return nil
	})
	printNode(state.Backing(), "1")
}
