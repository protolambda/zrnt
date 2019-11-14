package views

import (
	. "github.com/protolambda/zrnt/experimental/tree"
)

type SubtreeView struct {
	BackingNode Node
	depth       uint8
}

// Result will be nil if an error occurred.
func (sv *SubtreeView) Get(i uint64) (Node, error) {
	return sv.BackingNode.Getter(i, sv.depth)
}

// Result will be nil if an error occurred.
func (sv *SubtreeView) Set(i uint64, node Node) error {
	s, err := sv.BackingNode.Setter(i, sv.depth)
	if err != nil {
		return err
	}
	sv.BackingNode = s(node)
	return nil
}

func (sv *SubtreeView) Backing() Node {
	return sv.BackingNode
}
