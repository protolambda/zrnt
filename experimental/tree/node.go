package tree

// A link is called to rebind a value
type Link func(v Node)

// TODO: refactor these to use generalized indices as tree position.

type GetterInteraction interface {
	Getter(target uint64, depth uint8) (Node, error)
}

type SetterInteraction interface {
	Setter(target uint64, depth uint8) (Link, error)
}

type ExpandIntoInteraction interface {
	ExpandInto(target uint64, depth uint8) (Link, error)
}

type NodeInteraction interface {
	GetterInteraction
	SetterInteraction
	ExpandIntoInteraction
}

type Node interface {
	ComputeRoot(h HashFn) Root
}

type RebindableNode interface {
	Node
	Bind(bindingLink Link)
}

type ComplexNode interface {
	Node
	NodeInteraction
}

type ComplexRebindableNode interface {
	RebindableNode
	NodeInteraction
}
