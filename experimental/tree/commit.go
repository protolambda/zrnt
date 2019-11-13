package tree

import (
	"fmt"
	"log"
)

// An immutable (L, R) pair with a link to the holding node.
// If L or R changes, the link is used to bind a new (L, *R) or (*L, R) pair in the holding value.
type Commit struct {
	// TODO: instead of value + bool, it could also be a pointer (nil case == computed false).
	//  But more objects/indirection/allocations.
	Value    Root
	computed bool // true if Value is set to H(L, R)
	Link     Link
	Left     Node
	Right    Node
}

func (c *Commit) Bind(bindingLink Link) {
	c.Link = bindingLink
}

func (c *Commit) ComputeRoot(h HashFn) Root {
	if c.computed {
		return c.Value
	}
	if c.Left == nil || c.Right == nil {
		panic("invalid state, cannot have left without right")
	}
	c.Value = h(c.Left.ComputeRoot(h), c.Right.ComputeRoot(h))
	c.computed = true
	return c.Value
}

func (c *Commit) RebindLeft(v Node) {
	if c.Link == nil {
		log.Println("cannot rebind from unbound node!")
		return
	}
	c.Link(&Commit{
		Value:    Root{},
		computed: false,
		Link:     c.Link,
		Left:     v,
		Right:    c.Right,
	})
}

func (c *Commit) RebindRight(v Node) {
	if c.Link == nil {
		log.Println("cannot rebind from unbound node!")
		return
	}
	c.Link(&Commit{
		Value:    Root{},
		computed: false,
		Link:     c.Link,
		Left:     c.Left,
		Right:    v,
	})
}

func (c *Commit) Expand() {
	next := &Commit{
		Value:    Root{},
		computed: false,
		Link:     c.Link,
		Left:     nil,
		Right:    nil,
	}
	left := &Commit{
		Value:    Root{},
		computed: false,
		Link:     next.RebindLeft,
		Left:     nil,
		Right:    nil,
	}
	right := &Commit{
		Value:    Root{},
		computed: false,
		Link:     next.RebindRight,
		Left:     nil,
		Right:    nil,
	}
	next.Left = left
	next.Right = right
	c.Link(next)
}

// Unsafe! Modifies L and R, without triggering a rebind in the parent
func (c *Commit) ExpandInplaceTo(nodes []Node, depth uint8) {
	c.computed = false
	if depth == 0 {
		panic("invalid usage")
	}
	if depth == 1 {
		c.Left = nodes[0]
		if r, ok := nodes[0].(RebindableNode); ok {
			r.Bind(c.RebindLeft)
		}
		if len(nodes) > 1 {
			c.Right = nodes[1]
			if r, ok := nodes[1].(RebindableNode); ok {
				r.Bind(c.RebindRight)
			}
		} else {
			c.Right = &ZeroHashes[0]
		}
	} else {
		pivot := uint64(1) << depth
		c.Left = &Commit{
			Value:    Root{},
			computed: false,
			Link:     c.RebindLeft,
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
				Link:     c.RebindRight,
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
	if pivot := uint64(1) << depth; target < pivot {
		if c.Left == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no left node", target, depth)
		}
		if left, ok := c.Left.(GetterInteraction); ok {
			return left.Getter(target, depth-1)
		} else {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: left node has no GetterInteraction", target, depth)
		}
	} else {
		if c.Right == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no right node", target, depth)
		}
		if right, ok := c.Right.(GetterInteraction); ok {
			return right.Getter(target&^pivot, depth-1)
		} else {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: right node has no GetterInteraction", target, depth)
		}
	}
}

func (c *Commit) ExpandInto(target uint64, depth uint8) (Link, error) {
	if depth == 0 {
		return c.Link, nil
	}
	if depth == 1 {
		if target == 0 {
			return c.RebindLeft, nil
		}
		if target == 1 {
			return c.RebindRight, nil
		}
	}
	if pivot := uint64(1) << depth; target < pivot {
		if c.Left == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no left node", target, depth)
		}
		if left, ok := c.Left.(ExpandIntoInteraction); ok {
			return left.ExpandInto(target, depth-1)
		} else {
			// stop immediate propagation of rebinds during this Set call.
			var tmp Node

			startC := &Commit{
				Link: func(v Node) {
					tmp = v
				},
				Left:     &ZeroHashes[depth-2],
				Right:    &ZeroHashes[depth-2],
			}
			tmp = startC
			// Get the setter, recurse into the new node
			l, err := startC.ExpandInto(target, depth-1)

			newLeftC := tmp.(*Commit)
			// Now update the link to attach the updates of the new left node
			newLeftC.Link = c.RebindLeft
			// And attach as the left node
			c.RebindLeft(newLeftC)
			return l, err
		}
	} else {
		if c.Right == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no right node", target, depth)
		}
		if right, ok := c.Right.(ExpandIntoInteraction); ok {
			return right.ExpandInto(target&^pivot, depth-1)
		} else {
			// stop immediate propagation of rebinds during this Set call.
			var tmp Node

			startC := &Commit{
				Link: func(v Node) {
					tmp = v
				},
				Left:  &ZeroHashes[depth-1],
				Right: &ZeroHashes[depth-1],
			}
			tmp = startC
			// Get the setter, recurse into the new node
			l, err := startC.ExpandInto(target&^pivot, depth-1)

			newRightC := tmp.(*Commit)
			// Now update the link to attach the updates of the new right node
			newRightC.Link = c.RebindRight
			// And attach as the right node
			c.RebindRight(newRightC)
			return l, err
		}
	}
}

func (c *Commit) Setter(target uint64, depth uint8) (Link, error) {
	if depth == 0 {
		return c.Link, nil
	}
	if depth == 1 {
		if target == 0 {
			return c.RebindLeft, nil
		}
		if target == 1 {
			return c.RebindRight, nil
		}
	}
	if pivot := uint64(1) << depth; target < pivot {
		if c.Left == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no left node", target, depth)
		}
		if left, ok := c.Left.(SetterInteraction); ok {
			return left.Setter(target, depth-1)
		} else {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: left node has no SetterInteraction", target, depth)
		}
	} else {
		if c.Right == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no right node", target, depth)
		}
		if right, ok := c.Right.(SetterInteraction); ok {
			return right.Setter(target&^pivot, depth-1)
		} else {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: right node has no SetterInteraction", target, depth)
		}
	}
}

// Temporarily decouples the commit from its parent to scope modifications within the "work" callback.
// Then rebinds to the parent, and propagates changes if there were any.
func (c *Commit) Batch(work func()) {
	link := c.Link
	var next Node
	c.Link = func(v Node) {
		next = v
	}
	work()
	if next != nil {
		link(next)
	}
	c.Link = link
}
