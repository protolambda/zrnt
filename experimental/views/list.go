package views

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type ListView struct {
	SubtreeView
	limit uint64
}

type ListLength uint64

func (ll ListLength) MerkleRoot(h HashFn) (out Root) {
	binary.LittleEndian.PutUint64(out[:], uint64(ll))
	return
}

func List(limit uint64, nodes ...Node) (out *ListView) {
	elementCount := uint64(len(nodes))
	if elementCount > limit {
		// TODO: or add error return?
		return nil
	}
	depth := GetDepth(limit)

	mixin := ListLength(len(nodes))

	contents := &Commit{}
	contents.ExpandInplaceTo(nodes, depth)

	root := &Commit{Left: contents, Right: mixin}

	return &ListView{
		SubtreeView: SubtreeView{
			Backing: root,
			depth:   depth + 1, // 1 extra for length mix-in
		},
		limit: limit,
	}
}

// Takes the backing and wraps it to interpret it as a list
func NewListView(backing ComplexNode, limit uint64) *ListView {
	return &ListView{
		SubtreeView: SubtreeView{
			Backing: backing,
			depth:   GetDepth(limit),
		},
		limit: limit,
	}
}

// TODO: append/pop modify both contents as list-length: batching the changes would be good.

func (lv *ListView) Append(v Node) error {
	ll, err := lv.Length()
	if err != nil {
		return err
	}
	if ll >= lv.limit {
		return fmt.Errorf("list length is %d and appending would exceed the list limit %d", ll, lv.limit)
	}
	// Appending is done by setting the node at the index list-length. And expanding where necessary as it is being set.
	setLast, err := lv.SubtreeView.Backing.ExpandInto(ll, lv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to append an item")
	}
	// Append the item by setting the newly allocated last item to it.
	// Update the view to the new tree containing this item.
	next := setLast(v)
	if nextBacking, ok := next.(ComplexNode); ok {
		lv.Backing = nextBacking
	} else {
		return fmt.Errorf("new value %v is not a ComplexNode to update the view backing with", next)
	}
	// And update the list length
	setLength, err := lv.SubtreeView.Backing.Setter(1, 1)
	if err != nil {
		return err
	}
	next = setLength(ListLength(ll + 1))
	if nextBacking, ok := next.(ComplexNode); ok {
		lv.Backing = nextBacking
	} else {
		return fmt.Errorf("new value %v is not a ComplexNode to update the view backing with", next)
	}
	return nil
}

func (lv *ListView) Pop() error {
	ll, err := lv.Length()
	if err != nil {
		return err
	}
	if ll == 0 {
		return fmt.Errorf("list length is 0 and no item can be popped")
	}
	// Popping is done by setting the node at the index list-length. And expanding where necessary as it is being set.
	setLast, err := lv.SubtreeView.Backing.ExpandInto(ll, lv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to pop an item")
	}
	// Pop the item by setting it to the zero hash
	// Update the view to the new tree containing this item.
	next := setLast(&ZeroHashes[0])
	if nextBacking, ok := next.(ComplexNode); ok {
		lv.Backing = nextBacking
	} else {
		return fmt.Errorf("new value %v is not a ComplexNode to update the view backing with", next)
	}
	// And update the list length
	setLength, err := lv.SubtreeView.Backing.Setter(1, 1)
	if err != nil {
		return err
	}
	next = setLength(ListLength(ll - 1))
	if nextBacking, ok := next.(ComplexNode); ok {
		lv.Backing = nextBacking
	} else {
		return fmt.Errorf("new value %v is not a ComplexNode to update the view backing with", next)
	}
	return nil
}

// Use .SubtreeView.Get(i) to work with the tree and get explicit tree errors instead of nil result.
func (lv *ListView) Get(i uint64) Node {
	ll, err := lv.Length()
	if err != nil || i >= ll {
		return nil
	}
	v, _ := lv.SubtreeView.Get(i)
	return v
}

// Use .SubtreeView.Set(i) to work with the tree and get explicit tree errors instead of nil result.
func (lv *ListView) Set(i uint64, node Node) error {
	ll, err := lv.Length()
	if err != nil {
		return err
	}
	if i >= ll {
		return fmt.Errorf("cannot set item at element index %d, list only has %d elements", i, ll)
	}
	if i >= lv.limit {
		return fmt.Errorf("list has a an invalid length of %d and cannot set an element at index %d because of a limit of %d elements", ll, i, lv.limit)
	}
	return lv.SubtreeView.Set(i, node)
}

func (lv *ListView) Length() (uint64, error) {
	v, err := lv.SubtreeView.Backing.Getter(1, 1)
	if err != nil {
		return 0, err
	}
	ll, ok := v.(ListLength)
	if !ok {
		return 0, fmt.Errorf("cannot read node %v as list-length", v)
	}
	if uint64(ll) > lv.limit {
		return 0, fmt.Errorf("cannot read list length, length appears to be bigger than limit allows")
	}
	return uint64(ll), nil
}
