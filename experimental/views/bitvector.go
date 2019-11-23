package views

import (
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type BitVectorType struct {
	BitLength uint64
}

func (cd *BitVectorType) DefaultNode() Node {
	depth := GetDepth(cd.BottomNodeLength())
	inner := &Commit{}
	inner.ExpandInplaceDepth(&ZeroHashes[0], depth)
	return inner
}

func (cd *BitVectorType) ViewFromBacking(node Node) View {
	depth := GetDepth(cd.BottomNodeLength())
	return &BitVectorView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth,
		},
		BitVectorType: cd,
	}
}

func (cd *BitVectorType) BottomNodeLength() uint64 {
	return (cd.BitLength + 0xff) >> 8
}

func (cd *BitVectorType) New() *BitVectorView {
	return cd.ViewFromBacking(cd.DefaultNode()).(*BitVectorView)
}

type BitVectorView struct {
	SubtreeView
	*BitVectorType
}

func (cv *BitVectorView) ViewRoot(h HashFn) Root {
	return cv.BackingNode.MerkleRoot(h)
}

func (cv *BitVectorView) subviewNode(i uint64) (r *Root, subIndex uint8, err error) {
	bottomIndex, subIndex := i >> 8, uint8(i)
	v, err := cv.SubtreeView.Get(bottomIndex)
	if err != nil {
		return nil, 0, err
	}
	r, ok := v.(*Root)
	if !ok {
		return nil, 0, fmt.Errorf("bitvector bottom node is not a root, at index %d", i)
	}
	return r, subIndex, nil
}

func (cv *BitVectorView) Get(i uint64) (BoolView, error) {
	if i >= cv.BitLength {
		return false, fmt.Errorf("bitvector has bit length %d, cannot get bit index %d", cv.BitLength, i)
	}
	r, subIndex, err := cv.subviewNode(i)
	if err != nil {
		return false, err
	}
	return BoolType.BoolViewFromBitfieldBacking(r, subIndex), nil
}

func (cv *BitVectorView) Set(i uint64, view BoolView) error {
	if i >= cv.BitLength {
		return fmt.Errorf("cannot set item at element index %d, bitvector only has %d bits", i, cv.BitLength)
	}
	r, subIndex, err := cv.subviewNode(i)
	if err != nil {
		return err
	}
	return cv.SubtreeView.Set(i, view.BackingFromBitfieldBase(r, subIndex))
}
