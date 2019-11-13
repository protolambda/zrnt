package views

import (
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type SubtreeView struct {
	Backing ComplexNode
	depth uint8
}

// Result will be nil if an error occurred.
func (sv *SubtreeView) Get(i uint64) (Node, error) {
	return sv.Backing.Getter(i, sv.depth)
}

// Result will be nil if an error occurred.
func (sv *SubtreeView) Set(i uint64, node Node) error {
	s, err := sv.Backing.Setter(i, sv.depth)
	if err != nil {
		return err
	}
	next := s(node)
	if nextBacking, ok := next.(ComplexNode); ok {
		sv.Backing = nextBacking
		return nil
	} else {
		return fmt.Errorf("new value %v is not a ComplexNode to update the view backing with", next)
	}
}

