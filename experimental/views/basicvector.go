package views

import (
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type BasicVectorType struct {
	Length uint64
	ElementType BasicType
}

func (cd *BasicVectorType) DefaultNode() Node {
	depth := GetDepth(cd.BottomNodeCount())
	inner := &Commit{}
	inner.ExpandInplaceDepth(&ZeroHashes[0], depth)
	return inner
}

func (cd *BasicVectorType) ViewFromBacking(node Node) View {
	depth := GetDepth(cd.BottomNodeCount())
	return &BasicVectorView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth,
		},
		BasicVectorType: cd,
	}
}

func (cd *BasicVectorType) ElementsPerBottomNode() uint64 {
	return 32 / cd.ElementType.ByteLength()
}

func (cd *BasicVectorType) BottomNodeCount() uint64 {
	perNode := cd.ElementsPerBottomNode()
	return (cd.Length + perNode - 1) / perNode
}

func (cd *BasicVectorType) TranslateIndex(index uint64) (nodeIndex uint64, intraNodeIndex uint8) {
	perNode := cd.ElementsPerBottomNode()
	return index / perNode, uint8(index & (perNode - 1))
}

func (cd *BasicVectorType) New() *BitVectorView {
	return cd.ViewFromBacking(cd.DefaultNode()).(*BitVectorView)
}

type BasicVectorView struct {
	SubtreeView
	*BasicVectorType
}

func (cv *BasicVectorView) ViewRoot(h HashFn) Root {
	return cv.BackingNode.MerkleRoot(h)
}

// Use .SubtreeView.Get(i) to work with the tree and bypass typing.
func (cv *BasicVectorView) Get(i uint64) (SubView, error) {
	if i >= cv.Length {
		return nil, fmt.Errorf("basic vector has length %d, cannot get index %d", cv.Length, i)
	}
	bottomIndex, subIndex := cv.TranslateIndex(i)
	v, err := cv.SubtreeView.Get(bottomIndex)
	if err != nil {
		return nil, err
	}
	r, ok := v.(*Root)
	if !ok {
		return nil, fmt.Errorf("basic vector bottom node is not a root, cannot get subview from it at vector index %d", i)
	}
	return cv.ElementType.SubViewFromBacking(r, subIndex), nil
}

// Use .SubtreeView.Set(i, v) to work with the tree and bypass typing.
func (cv *BasicVectorView) Set(i uint64, view SubView) error {
	if i >= cv.Length {
		return fmt.Errorf("cannot set item at element index %d, basic vector only has %d elements", i, cv.Length)
	}
	bottomIndex, subIndex := cv.TranslateIndex(i)
	v, err := cv.SubtreeView.Get(bottomIndex)
	if err != nil {
		return err
	}
	r, ok := v.(*Root)
	if !ok {
		return fmt.Errorf("basic vector bottom node is not a root, cannot set subview from it at vector index %d", i)
	}
	return cv.SubtreeView.Set(i, view.BackingFromBase(r, subIndex))
}
