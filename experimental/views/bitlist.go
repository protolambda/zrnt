package views

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
)

type BitListType struct {
	BitLimit uint64
}

func (cd *BitListType) DefaultNode() Node {
	depth := GetDepth(cd.BottomNodeLimit())
	return &Commit{Left: &ZeroHashes[depth], Right: &ZeroHashes[0]}
}

func (cd *BitListType) ViewFromBacking(node Node) View {
	depth := GetDepth(cd.BottomNodeLimit())
	return &BitListView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth + 1, // +1 for length mix-in
		},
		BitListType: cd,
	}
}

func (cd *BitListType) BottomNodeLimit() uint64 {
	return (cd.BitLimit + 0xff) >> 8
}

func (cd *BitListType) New() *BitListView {
	return cd.ViewFromBacking(cd.DefaultNode()).(*BitListView)
}

type BitListView struct {
	SubtreeView
	*BitListType
}

func (cv *BitListView) ViewRoot(h HashFn) Root {
	return cv.BackingNode.MerkleRoot(h)
}

func (cv *BitListView) Append(view BoolView) error {
	ll, err := cv.Length()
	if err != nil {
		return err
	}
	if ll >= cv.BitLimit {
		return fmt.Errorf("list length is %d and appending would exceed the list limit %d", ll, cv.BitLimit)
	}
	// Appending is done by modifying the bottom node at the index list_length. And expanding where necessary as it is being set.
	setLast, err := cv.SubtreeView.BackingNode.ExpandInto(ll >> 8, cv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to append an item")
	}
	if ll & 0xff == 0 {
		// New bottom node
		cv.BackingNode = setLast(view.BackingFromBitfieldBase(&ZeroHashes[0], 0))
	} else {
		// Apply to existing partially zeroed bottom node
		r, subIndex, err := cv.subviewNode(ll)
		if err != nil {
			return err
		}
		cv.BackingNode = setLast(view.BackingFromBitfieldBase(r, subIndex))
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

func (cv *BitListView) Pop() error {
	ll, err := cv.Length()
	if err != nil {
		return err
	}
	if ll == 0 {
		return fmt.Errorf("list length is 0 and no bit can be popped")
	}
	// Popping is done by modifying the bottom node at the index list_length - 1. And expanding where necessary as it is being set.
	setLast, err := cv.SubtreeView.BackingNode.ExpandInto((ll-1) >> 8, cv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to pop a bit")
	}
	// Get the subview to erase
	r, subIndex, err := cv.subviewNode(ll - 1)
	if err != nil {
		return err
	}
	// Pop the bit by setting it to the default
	// Update the view to the new tree containing this item.
	cv.BackingNode = setLast(BoolView(false).BackingFromBitfieldBase(r, subIndex))
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

func (lv *BitListView) CheckIndex(i uint64) error {
	ll, err := lv.Length()
	if err != nil {
		return err
	}
	if i >= ll {
		return fmt.Errorf("cannot handle item at element index %d, list only has %d bits", i, ll)
	}
	if i >= lv.BitLimit {
		return fmt.Errorf("bitlist has a an invalid length of %d and cannot handle a bit at index %d because of a limit of %d bits", ll, i, lv.BitLimit)
	}
	return nil
}

func (cv *BitListView) subviewNode(i uint64) (r *Root, subIndex uint8, err error) {
	bottomIndex, subIndex := i >> 8, uint8(i)
	v, err := cv.SubtreeView.Get(bottomIndex)
	if err != nil {
		return nil, 0, err
	}
	r, ok := v.(*Root)
	if !ok {
		return nil, 0, fmt.Errorf("bitlist bottom node is not a root, at index %d", i)
	}
	return r, subIndex, nil
}

func (cv *BitListView) Get(i uint64) (BoolView, error) {
	if err := cv.CheckIndex(i); err != nil {
		return false, err
	}
	r, subIndex, err := cv.subviewNode(i)
	if err != nil {
		return false, err
	}
	return BoolType.BoolViewFromBitfieldBacking(r, subIndex), nil
}

func (cv *BitListView) Set(i uint64, view BoolView) error {
	if err := cv.CheckIndex(i); err != nil {
		return err
	}
	r, subIndex, err := cv.subviewNode(i)
	if err != nil {
		return err
	}
	return cv.SubtreeView.Set(i, view.BackingFromBitfieldBase(r, subIndex))
}

func (lv *BitListView) Length() (uint64, error) {
	v, err := lv.SubtreeView.BackingNode.Getter(1, 1)
	if err != nil {
		return 0, err
	}
	llBytes, ok := v.(*Root)
	if !ok {
		return 0, fmt.Errorf("cannot read node %v as list-length", v)
	}
	ll := binary.LittleEndian.Uint64(llBytes[:8])
	if ll > lv.BitLimit {
		return 0, fmt.Errorf("cannot read list length, length appears to be bigger than limit allows")
	}
	return ll, nil
}
