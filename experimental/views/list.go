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

func (ll ListLength) ComputeRoot(h HashFn) (out Root) {
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
	out = &ListView{
		SubtreeView: SubtreeView{
			depth: depth + 1,
		},
		limit: limit,
	} // 1 extra for length mix-in

	mixin := ListLength(len(nodes))

	contents := &Commit{}
	contents.ExpandInplaceTo(nodes, depth)

	root := &Commit{Link: out.RebindInner, Left: contents, Right: mixin}
	contents.Bind(root.RebindLeft)

	out.ComplexNode = root
	return out
}

// Takes the backing, binds the view as a proxy link to keep track of changes,
// and defines a List based on a limit and a dynamic number of number of elements.
// The link however is optional: if nil, no changes are propagated upwards,
// but the view itself still tracks the latest changes.
func NewListView(backing ComplexRebindableNode, link Link, limit uint64) *ListView {
	view := &ListView{
		SubtreeView: SubtreeView{
			ComplexNode: backing,
			Link:        link,
			depth:       GetDepth(limit),
		},
		limit: limit,
	}
	backing.Bind(view.RebindInner)
	return view
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
	setLast, err := lv.ExpandInto(ll, lv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to append an item")
	}
	// Append the item by setting the newly allocated last item to it.
	setLast(v)
	// And update the list length
	setLength, err := lv.SubtreeView.Setter(1, 1)
	if err != nil {
		return err
	}
	setLength(ListLength(ll + 1))
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
	setLast, err := lv.ExpandInto(ll, lv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to append an item")
	}
	// Pop the item by setting it to the zero hash
	setLast(&ZeroHashes[0])
	// And update the list length
	setLength, err := lv.SubtreeView.Setter(1, 1)
	if err != nil {
		return err
	}
	setLength(ListLength(ll - 1))
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
func (lv *ListView) Set(i uint64) Link {
	ll, err := lv.Length()
	if err != nil || i >= ll {
		return nil
	}
	v, _ := lv.SubtreeView.Set(i)
	return v
}

func (lv *ListView) Length() (uint64, error) {
	v, err := lv.SubtreeView.Getter(1, 1)
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
