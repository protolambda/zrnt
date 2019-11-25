package tree

import "fmt"

type Root [32]byte

// Backing, a root can be used as a view representing itself.
func (s *Root) Backing() Node {
	return s
}

func (s *Root) Getter(target uint64, depth uint8) (Node, error) {
	if depth != 0 {
		return nil, fmt.Errorf("A Root does not have any child nodes to Get")
	}
	return s, nil
}

func (s *Root) Setter(target uint64, depth uint8) (Link, error) {
	if depth != 0 {
		return nil, fmt.Errorf("A Root does not have any child nodes to Set")
	}
	return Identity, nil
}

func (s *Root) ExpandInto(target uint64, depth uint8) (Link, error) {
	if depth == 0 {
		return Identity, nil
	}
	startC := &Commit{
		Left:  &ZeroHashes[depth-1],
		Right: &ZeroHashes[depth-1],
	}
	return startC.ExpandInto(target, depth-1)
}

func (s *Root) MerkleRoot(h HashFn) Root {
	if s == nil {
		return Root{}
	}
	return *s
}

type RootMeta uint8

func (RootMeta) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (RootMeta) ViewFromBacking(node Node) View {
	root, ok := node.(*Root)
	if !ok {
		return nil
	} else {
		return root
	}
}

const RootType RootMeta = 0
