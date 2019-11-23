package views

import (
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type VectorType struct {
	ElementType TypeDef
	Length uint64
}

func (cd *VectorType) DefaultNode() Node {
	depth := GetDepth(cd.Length)
	inner := &Commit{}
	// The same node N times: the node is immutable, so re-use is safe.
	defaultNode := cd.ElementType.DefaultNode()
	inner.ExpandInplaceDepth(defaultNode, depth)
	return inner
}

func (cd *VectorType) ViewFromBacking(node Node) View {
	depth := GetDepth(cd.Length)
	return &VectorView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth,
		},
		VectorType: cd,
	}
}

func (cd *VectorType) New() *VectorView {
	return cd.ViewFromBacking(cd.DefaultNode()).(*VectorView)
}

type VectorView struct {
	SubtreeView
	*VectorType
}

func (cv *VectorView) ViewRoot(h HashFn) Root {
	return cv.BackingNode.MerkleRoot(h)
}

func (cv *VectorView) Get(i uint64) (View, error) {
	if i >= cv.VectorType.Length {
		return nil, fmt.Errorf("cannot get item at element index %d, vector only has %d elements", i, cv.VectorType.Length)
	}
	v, err := cv.SubtreeView.Get(i)
	if err != nil {
		return nil, err
	}
	return cv.VectorType.ElementType.ViewFromBacking(v), nil
}

func (cv *VectorView) Set(i uint64, view View) error {
	if i >= cv.VectorType.Length {
		return fmt.Errorf("cannot set item at element index %d, vector only has %d elements", i, cv.VectorType.Length)
	}
	return cv.SubtreeView.Set(i, view.Backing())
}
