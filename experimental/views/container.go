package views

import (
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type ContainerType struct {
	Fields []TypeDef
}

func (cd *ContainerType) DefaultNode() Node {
	fieldCount := cd.FieldCount()
	depth := GetDepth(fieldCount)
	inner := &Commit{}
	nodes := make([]Node, fieldCount, fieldCount)
	for i, f := range cd.Fields {
		nodes[i] = f.DefaultNode()
	}
	inner.ExpandInplaceTo(nodes, depth)
	return inner
}

func (cd *ContainerType) ViewFromBacking(node Node) View {
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

func (cd *ContainerType) New() *ContainerView {
	return cd.ViewFromBacking(cd.DefaultNode()).(*ContainerView)
}

func (cd *ContainerType) FieldCount() uint64 {
	return uint64(len(cd.Fields))
}

type ContainerView struct {
	SubtreeView
	*ContainerType
}

func (cv *ContainerView) ViewRoot(h HashFn) Root {
	return cv.BackingNode.MerkleRoot(h)
}

// Use .SubtreeView.Get(i) to work with the tree and get explicit tree errors instead of nil result.
func (cv *ContainerView) Get(i uint64) View {
	if i >= cv.ContainerType.FieldCount() {
		return nil
	}
	v, err := cv.SubtreeView.Get(i)
	if err != nil {
		return nil
	}
	return cv.ContainerType.Fields[i].ViewFromBacking(v)
}

// Use .SubtreeView.Set(i, v) to work with the tree and bypass typing.
func (cv *ContainerView) Set(i uint64, view View) error {
	if fieldCount := cv.ContainerType.FieldCount(); i >= fieldCount {
		return fmt.Errorf("cannot set item at field index %d, container only has %d fields", i, fieldCount)
	}
	return cv.SubtreeView.Set(i, view.Backing())
}
