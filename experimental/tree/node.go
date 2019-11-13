package tree

// A link is called to rebind a value, and retrieve the new root node.
type Link func(v Node) Node

func Identity(v Node) Node {
	return v
}

func Compose(inner Link, outer Link) Link {
	return func(v Node) Node {
		return outer(inner(v))
	}
}

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
	MerkleRoot(h HashFn) Root
}

type ComplexNode interface {
	Node
	NodeInteraction
}
