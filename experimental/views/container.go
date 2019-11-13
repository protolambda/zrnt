package views

import (
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type ContainerView struct {
	SubtreeView
	fieldCount uint64
}

func Container(nodes ...Node) (out *ContainerView) {
	elementCount := uint64(len(nodes))
	depth := GetDepth(elementCount)
	inner := &Commit{}
	inner.ExpandInplaceTo(nodes, depth)
	out = &ContainerView{
		SubtreeView: SubtreeView{
			Backing: inner,
			depth:   depth,
		},
		fieldCount: elementCount,
	}
	return out
}

// Takes the backing and wraps it to interpret it as a container
func NewContainerView(backing ComplexNode, elementCount uint64) *ContainerView {
	return &ContainerView{
		SubtreeView: SubtreeView{
			Backing: backing,
			depth:   GetDepth(elementCount),
		},
		fieldCount: elementCount,
	}
}

func (cv *ContainerView) ViewRoot(h HashFn) Root {
	return cv.Backing.MerkleRoot(h)
}

// Use .SubtreeView.Get(i) to work with the tree and get explicit tree errors instead of nil result.
func (cv *ContainerView) Get(i uint64) Node {
	if i >= cv.fieldCount {
		return nil
	}
	v, _ := cv.SubtreeView.Get(i)
	return v
}

// Use .SubtreeView.Set(i, v) to work with the tree and bypass typing.
func (cv *ContainerView) Set(i uint64, node Node) error {
	if i >= cv.fieldCount {
		return fmt.Errorf("cannot set item at field index %d, container only has %d fields", i, cv.fieldCount)
	}
	return cv.SubtreeView.Set(i, node)
}

func (cv *ContainerView) Length() uint64 {
	return cv.fieldCount
}
