package tree

import (
	"fmt"
)

// An immutable (L, R) pair with a link to the holding node.
// If L or R changes, the link is used to bind a new (L, *R) or (*L, R) pair in the holding value.
type Commit struct {
	// TODO: instead of value + bool, it could also be a pointer (nil case == computed false).
	//  But more objects/indirection/allocations.
	Value    Root
	computed bool // true if Value is set to H(L, R)
	Left     Node
	Right    Node
}

func (c *Commit) MerkleRoot(h HashFn) Root {
	if c.computed {
		return c.Value
	}
	if c.Left == nil || c.Right == nil {
		panic("invalid state, cannot have left without right")
	}
	c.Value = h(c.Left.MerkleRoot(h), c.Right.MerkleRoot(h))
	c.computed = true
	return c.Value
}

func (c *Commit) RebindLeft(v Node) Node {
	return &Commit{
		Value:    Root{},
		computed: false,
		Left:     v,
		Right:    c.Right,
	}
}

func (c *Commit) RebindRight(v Node) Node {
	return &Commit{
		Value:    Root{},
		computed: false,
		Left:     c.Left,
		Right:    v,
	}
}

func (c *Commit) Expand() Node {
	next := &Commit{
		Value:    Root{},
		computed: false,
		Left:     nil,
		Right:    nil,
	}
	left := &Commit{
		Value:    Root{},
		computed: false,
		Left:     nil,
		Right:    nil,
	}
	right := &Commit{
		Value:    Root{},
		computed: false,
		Left:     nil,
		Right:    nil,
	}
	next.Left = left
	next.Right = right
	return next
}

// Unsafe! Modifies L and R, without triggering a rebind in the parent
func (c *Commit) ExpandInplaceTo(nodes []Node, depth uint8) {
	c.computed = false
	if depth == 0 {
		panic("invalid usage")
	}
	if depth == 1 {
		c.Left = nodes[0]
		if len(nodes) > 1 {
			c.Right = nodes[1]
		} else {
			c.Right = &ZeroHashes[0]
		}
	} else {
		pivot := uint64(1) << (depth - 1)
		c.Left = &Commit{
			Value:    Root{},
			computed: false,
			Left:     nil,
			Right:    nil,
		}
		if uint64(len(nodes)) <= pivot {
			c.Left.(*Commit).ExpandInplaceTo(nodes, depth-1)
			c.Right = &ZeroHashes[depth]
		} else {
			c.Left.(*Commit).ExpandInplaceTo(nodes[:pivot], depth-1)
			c.Right = &Commit{
				Value:    Root{},
				computed: false,
				Left:     nil,
				Right:    nil,
			}
			c.Right.(*Commit).ExpandInplaceTo(nodes[pivot:], depth-1)
		}
	}
}

func (c *Commit) Getter(target uint64, depth uint8) (Node, error) {
	if depth == 0 {
		return c, nil
	}
	if depth == 1 {
		if target == 0 {
			return c.Left, nil
		}
		if target == 1 {
			return c.Right, nil
		}
	}
	if pivot := uint64(1) << (depth - 1); target < pivot {
		if c.Left == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no left node", target, depth)
		}
		return c.Left.Getter(target, depth-1)
	} else {
		if c.Right == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no right node", target, depth)
		}
		return c.Right.Getter(target&^pivot, depth-1)
	}
}

func (c *Commit) ExpandInto(target uint64, depth uint8) (Link, error) {
	if depth == 0 {
		return Identity, nil
	}
	if depth == 1 {
		if target == 0 {
			return c.RebindLeft, nil
		}
		if target == 1 {
			return c.RebindRight, nil
		}
	}
	if pivot := uint64(1) << (depth - 1); target < pivot {
		if c.Left == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no left node", target, depth)
		}
		if inner, err := c.Left.ExpandInto(target, depth-1); err != nil {
			return nil, err
		} else {
			return Compose(inner, c.RebindLeft), nil
		}
	} else {
		if c.Right == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no right node", target, depth)
		}
		if inner, err := c.Right.ExpandInto(target&^pivot, depth-1); err != nil {
			return nil, err
		} else {
			return Compose(inner, c.RebindRight), nil
		}
	}
}

func (c *Commit) Setter(target uint64, depth uint8) (Link, error) {
	if depth == 0 {
		return Identity, nil
	}
	if depth == 1 {
		if target == 0 {
			return c.RebindLeft, nil
		}
		if target == 1 {
			return c.RebindRight, nil
		}
	}
	if pivot := uint64(1) << (depth - 1); target < pivot {
		if c.Left == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no left node", target, depth)
		}
		if inner, err := c.Left.Setter(target, depth-1); err != nil {
			return nil, err
		} else {
			return Compose(inner, c.RebindLeft), nil
		}
	} else {
		if c.Right == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no right node", target, depth)
		}
		if inner, err := c.Right.Setter(target&^pivot, depth-1); err != nil {
			return nil, err
		} else {
			return Compose(inner, c.RebindRight), nil
		}
	}
}

// TODO: do we need a batching pattern, to not rebind branch by branch? Or is it sufficient to only create setters with reasonable scope?
