package views

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type BasicListType struct {
	ElementType BasicTypeDef
	Limit       uint64
}

func (cd *BasicListType) DefaultNode() Node {
	depth := GetDepth(cd.BottomNodeLimit())
	return &Commit{Left: &ZeroHashes[depth], Right: &ZeroHashes[0]}
}

func (cd *BasicListType) ViewFromBacking(node Node) View {
	depth := GetDepth(cd.BottomNodeLimit())
	return &BasicListView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth + 1, // +1 for length mix-in
		},
		BasicListType: cd,
	}
}

func (cd *BasicListType) ElementsPerBottomNode() uint64 {
	return 32 / cd.ElementType.ByteLength()
}

func (cd *BasicListType) BottomNodeLimit() uint64 {
	perNode := cd.ElementsPerBottomNode()
	return (cd.Limit + perNode - 1) / perNode
}

func (cd *BasicListType) TranslateIndex(index uint64) (nodeIndex uint64, intraNodeIndex uint8) {
	perNode := cd.ElementsPerBottomNode()
	return index / perNode, uint8(index & (perNode - 1))
}

func (cd *BasicListType) New() *BasicListView {
	return cd.ViewFromBacking(cd.DefaultNode()).(*BasicListView)
}

type BasicListView struct {
	SubtreeView
	*BasicListType
}

func (cv *BasicListView) ViewRoot(h HashFn) Root {
	return cv.BackingNode.MerkleRoot(h)
}

func (cv *BasicListView) Append(view SubView) error {
	ll, err := cv.Length()
	if err != nil {
		return err
	}
	if ll >= cv.Limit {
		return fmt.Errorf("list length is %d and appending would exceed the list limit %d", ll, cv.Limit)
	}
	perNode := cv.ElementsPerBottomNode()
	// Appending is done by modifying the bottom node at the index list_length. And expanding where necessary as it is being set.
	setLast, err := cv.SubtreeView.BackingNode.ExpandInto(ll/perNode, cv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to append an item")
	}
	if ll%perNode == 0 {
		// New bottom node
		cv.BackingNode = setLast(view.BackingFromBase(&ZeroHashes[0], 0))
	} else {
		// Apply to existing partially zeroed bottom node
		r, _, subIndex, err := cv.subviewNode(ll)
		if err != nil {
			return err
		}
		cv.BackingNode = setLast(view.BackingFromBase(r, subIndex))
	}
	// And update the list length
	setLength, err := cv.SubtreeView.BackingNode.Setter(1, 1)
	if err != nil {
		return err
	}
	newLength := &Root{}
	binary.LittleEndian.PutUint64(newLength[:8], ll+1)
	cv.BackingNode = setLength(newLength)
	return nil
}

func (cv *BasicListView) Pop() error {
	ll, err := cv.Length()
	if err != nil {
		return err
	}
	if ll == 0 {
		return fmt.Errorf("list length is 0 and no item can be popped")
	}
	perNode := cv.ElementsPerBottomNode()
	// Popping is done by modifying the bottom node at the index list_length - 1. And expanding where necessary as it is being set.
	setLast, err := cv.SubtreeView.BackingNode.ExpandInto((ll-1)/perNode, cv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to pop an item")
	}
	// Get the subview to erase
	r, _, subIndex, err := cv.subviewNode(ll - 1)
	if err != nil {
		return err
	}
	// Pop the item by setting it to the default
	// Update the view to the new tree containing this item.
	defaultElement := cv.ElementType.ViewFromBacking(cv.ElementType.DefaultNode()).(SubView)
	cv.BackingNode = setLast(defaultElement.BackingFromBase(r, subIndex))
	// And update the list length
	setLength, err := cv.SubtreeView.BackingNode.Setter(1, 1)
	if err != nil {
		return err
	}
	newLength := &Root{}
	binary.LittleEndian.PutUint64(newLength[:8], ll-1)
	cv.BackingNode = setLength(newLength)
	return nil
}

func (lv *BasicListView) CheckIndex(i uint64) error {
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

func (cv *BasicListView) subviewNode(i uint64) (r *Root, bottomIndex uint64, subIndex uint8, err error) {
	bottomIndex, subIndex = cv.TranslateIndex(i)
	v, err := cv.SubtreeView.Get(bottomIndex)
	if err != nil {
		return nil, 0, 0, err
	}
	r, ok := v.(*Root)
	if !ok {
		return nil, 0, 0, fmt.Errorf("basic list bottom node is not a root, at index %d", i)
	}
	return r, bottomIndex, subIndex, nil
}

func (cv *BasicListView) Get(i uint64) (SubView, error) {
	if err := cv.CheckIndex(i); err != nil {
		return nil, err
	}
	r, _, subIndex, err := cv.subviewNode(i)
	if err != nil {
		return nil, err
	}
	return cv.ElementType.SubViewFromBacking(r, subIndex), nil
}

func (cv *BasicListView) Set(i uint64, view SubView) error {
	if err := cv.CheckIndex(i); err != nil {
		return err
	}
	r, bottomIndex, subIndex, err := cv.subviewNode(i)
	if err != nil {
		return err
	}
	return cv.SubtreeView.Set(bottomIndex, view.BackingFromBase(r, subIndex))
}

func (lv *BasicListView) Length() (uint64, error) {
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
