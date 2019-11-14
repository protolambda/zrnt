package tree

import "fmt"

type Root [32]byte

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
