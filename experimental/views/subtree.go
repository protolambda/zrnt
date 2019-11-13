package views

import . "github.com/protolambda/zrnt/experimental/tree"

type SubtreeView struct {
	ComplexNode
	Link
	depth uint8
}

func (sv *SubtreeView) Bind(bindingLink Link) {
	sv.Link = bindingLink
}

func (sv *SubtreeView) RebindInner(v Node) {
	// the view keeps track of the latest node, but proxies the rebind up without stepping in-between.
	sv.ComplexNode = v.(ComplexNode)

	if sv.Link == nil {
		// If nil, the view is maintaining the latest root binding
		return
	}
	sv.Link(v)
}

// Result will be nil if an error occurred.
func (sv *SubtreeView) Get(i uint64) (Node, error) {
	return sv.Getter(i, sv.depth)
}

// Result will be nil if an error occurred.
func (sv *SubtreeView) Set(i uint64) (Link, error) {
	return sv.Setter(i, sv.depth)
}

