package views

import (
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type ContainerType []TypeDef

func (cd ContainerType) DefaultNode() Node {
	fieldCount := cd.FieldCount()
	depth := GetDepth(fieldCount)
	inner := &Commit{}
	nodes := make([]Node, fieldCount, fieldCount)
	for i, f := range cd {
		nodes[i] = f.DefaultNode()
	}
	inner.ExpandInplaceTo(nodes, depth)
	return inner
}

func (cd ContainerType) ViewFromBacking(node Node) View {
	fieldCount := cd.FieldCount()
	depth := GetDepth(fieldCount)
	return &ContainerView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth,
		},
		ContainerType: cd,
	}
}

func (cd ContainerType) New() *ContainerView {
	return cd.ViewFromBacking(cd.DefaultNode()).(*ContainerView)
}

func (cd ContainerType) FieldCount() uint64 {
	return uint64(len(cd))
}

type ContainerView struct {
	SubtreeView
	ContainerType
}

func (cv *ContainerView) ViewRoot(h HashFn) Root {
	return cv.BackingNode.MerkleRoot(h)
}

func (cv *ContainerView) Get(i uint64) (View, error) {
	if count := cv.ContainerType.FieldCount(); i >= count {
		return nil, fmt.Errorf("cannot get item at field index %d, container only has %d fields", i, count)
	}
	v, err := cv.SubtreeView.Get(i)
	if err != nil {
		return nil, err
	}
	return cv.ContainerType[i].ViewFromBacking(v), nil
}

func (cv *ContainerView) Set(i uint64, view View) error {
	if fieldCount := cv.ContainerType.FieldCount(); i >= fieldCount {
		return fmt.Errorf("cannot set item at field index %d, container only has %d fields", i, fieldCount)
	}
	return cv.SubtreeView.Set(i, view.Backing())
}
