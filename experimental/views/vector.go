package views

import . "github.com/protolambda/zrnt/experimental/tree"

type VectorView struct {
	SubtreeView
	length uint64
}

func Vector(length uint64, nodeFn func(i uint64) Node) (out *VectorView) {
	depth := GetDepth(length)
	out = &VectorView{
		SubtreeView: SubtreeView{
			depth: depth,
		},
		length: length,
	}
	inner := &Commit{Link: out.RebindInner}
	nodes := make([]Node, length, length)
	for i := uint64(0); i < length; i++ {
		nodes[i] = nodeFn(i)
	}
	inner.ExpandInplaceTo(nodes, depth)
	out.ComplexNode = inner
	return out
}

// Takes the backing, binds the view as a proxy link to keep track of changes,
// and defines a Vector based on a fixed number of elements.
// The link however is optional: if nil, no changes are propagated upwards,
// but the view itself still tracks the latest changes.
func NewVectorView(backing ComplexRebindableNode, link Link, length uint64) *VectorView {
	view := &VectorView{
		SubtreeView: SubtreeView{
			ComplexNode: backing,
			Link:        link,
			depth:       GetDepth(length),
		},
		length: length,
	}
	backing.Bind(view.RebindInner)
	return view
}

// Use .SubtreeView.Get(i) to work with the tree and get explicit tree errors instead of nil result.
func (vv *VectorView) Get(i uint64) Node {
	if i >= vv.length {
		return nil
	}
	v, _ := vv.SubtreeView.Get(i)
	return v
}

// Use .SubtreeView.Set(i) to work with the tree and get explicit tree errors instead of nil result.
func (vv *VectorView) Set(i uint64) Link {
	if i >= vv.length {
		return nil
	}
	v, _ := vv.SubtreeView.Set(i)
	return v
}

func (vv *VectorView) Length() uint64 {
	return vv.length
}
