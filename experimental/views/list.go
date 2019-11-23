package views

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type ListType struct {
	ElementType TypeDef
	Limit       uint64
}

func (cd *ListType) DefaultNode() Node {
	depth := GetDepth(cd.Limit)
	return &Commit{Left: &ZeroHashes[depth], Right: &ZeroHashes[0]}
}

func (cd *ListType) ViewFromBacking(node Node) View {
	depth := GetDepth(cd.Limit)
	return &ListView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth + 1, // +1 for length mix-in
		},
		ListType: cd,
	}
}

func (cd *ListType) New() *ListView {
	return cd.ViewFromBacking(cd.DefaultNode()).(*ListView)
}

type ListView struct {
	SubtreeView
	*ListType
}

func (lv *ListView) ViewRoot(h HashFn) Root {
	return lv.BackingNode.MerkleRoot(h)
}

func (lv *ListView) Append(v Node) error {
	ll, err := lv.Length()
	if err != nil {
		return err
	}
	if ll >= lv.Limit {
		return fmt.Errorf("list length is %d and appending would exceed the list limit %d", ll, lv.Limit)
	}
	// Appending is done by setting the node at the index list_length. And expanding where necessary as it is being set.
	setLast, err := lv.SubtreeView.BackingNode.ExpandInto(ll, lv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to append an item")
	}
	// Append the item by setting the newly allocated last item to it.
	// Update the view to the new tree containing this item.
	lv.BackingNode = setLast(v)
	// And update the list length
	setLength, err := lv.SubtreeView.BackingNode.Setter(1, 1)
	if err != nil {
		return err
	}
	newLength := &Root{}
	binary.LittleEndian.PutUint64(newLength[:8], ll+1)
	lv.BackingNode = setLength(newLength)
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
	// Popping is done by setting the node at the index list_length - 1. And expanding where necessary as it is being set.
	setLast, err := lv.SubtreeView.BackingNode.ExpandInto(ll-1, lv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to pop an item")
	}
	// Pop the item by setting it to the zero hash
	// Update the view to the new tree containing this item.
	lv.BackingNode = setLast(&ZeroHashes[0])
	// And update the list length
	setLength, err := lv.SubtreeView.BackingNode.Setter(1, 1)
	if err != nil {
		return err
	}
	newLength := &Root{}
	binary.LittleEndian.PutUint64(newLength[:8], ll-1)
	lv.BackingNode = setLength(newLength)
	return nil
}

func (lv *ListView) CheckIndex(i uint64) error {
	ll, err := lv.Length()
	if err != nil {
		return err
	}
	if i >= ll {
		return fmt.Errorf("cannot handle item at element index %d, list only has %d elements", i, ll)
	}
	if i >= lv.Limit {
		return fmt.Errorf("list has a an invalid length of %d and cannot handle an element at index %d because of a limit of %d elements", ll, i, lv.Limit)
	}
	return nil
}

func (lv *ListView) Get(i uint64) (Node, error) {
	if err := lv.CheckIndex(i); err != nil {
		return nil, err
	}
	return lv.SubtreeView.Get(i)
}

func (lv *ListView) Set(i uint64, node Node) error {
	if err := lv.CheckIndex(i); err != nil {
		return err
	}
	return lv.SubtreeView.Set(i, node)
}

func (lv *ListView) Length() (uint64, error) {
	v, err := lv.SubtreeView.BackingNode.Getter(1, 1)
	if err != nil {
		return 0, err
	}
	llBytes, ok := v.(*Root)
	if !ok {
		return 0, fmt.Errorf("cannot read node %v as list-length", v)
	}
	ll := binary.LittleEndian.Uint64(llBytes[:8])
	if ll > lv.Limit {
		return 0, fmt.Errorf("cannot read list length, length appears to be bigger than limit allows")
	}
	return ll, nil
}
