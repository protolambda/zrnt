package views

import . "github.com/protolambda/zrnt/experimental/tree"

type ContainerView struct {
	SubtreeView
	fieldCount uint64
}

func Container(nodes ...Node) (out *ContainerView) {
	elementCount := uint64(len(nodes))
	depth := GetDepth(elementCount)
	out = &ContainerView{
		SubtreeView: SubtreeView{
			depth: depth,
		},
		fieldCount: elementCount,
	}
	inner := &Commit{Link: out.RebindInner}
	inner.ExpandInplaceTo(nodes, depth)
	out.ComplexNode = inner
	return out
}

// Takes the backing, binds the view as a proxy link to keep track of changes,
// and defines a Container based on a fixed number of elements.
// The link however is optional: if nil, no changes are propagated upwards,
// but the view itself still tracks the latest changes.
func NewContainerView(backing ComplexRebindableNode, link Link, elementCount uint64) *ContainerView {
	view := &ContainerView{
		SubtreeView: SubtreeView{
			ComplexNode: backing,
			Link:        link,
			depth:       GetDepth(elementCount),
		},
		fieldCount: elementCount,
	}
	backing.Bind(view.RebindInner)
	return view
}

// Use .SubtreeView.Get(i) to work with the tree and get explicit tree errors instead of nil result.
func (cv *ContainerView) Get(i uint64) Node {
	if i >= cv.fieldCount {
		return nil
	}
	v, _ := cv.SubtreeView.Get(i)
	return v
}

// Use .SubtreeView.Set(i) to work with the tree and get explicit tree errors instead of nil result.
func (cv *ContainerView) Set(i uint64) Link {
	if i >= cv.fieldCount {
		return nil
	}
	v, _ := cv.SubtreeView.Set(i)
	return v
}

func (cv *ContainerView) Length() uint64 {
	return cv.fieldCount
}
