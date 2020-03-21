package header

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var SignedBeaconBlockHeaderType = &ContainerType{
	{"message", BeaconBlockHeaderType},
	{"signature", BLSSignatureType},
}

var BeaconBlockHeaderType = &ContainerType{
	{"slot", SlotType},
	{"proposer_index", ValidatorIndexType},
	{"parent_root", RootType},
	{"state_root", RootType},
	{"body_root", RootType},
}

type BeaconBlockHeaderNode struct { *ContainerView }

func NewBeaconBlockHeaderNode() *BeaconBlockHeaderNode {
	return &BeaconBlockHeaderNode{ContainerView: BeaconBlockHeaderType.New(nil)}
}

func (v *BeaconBlockHeaderNode) HashTreeRoot() Root {
	return v.ViewRoot(tree.Hash)
}
func (v *BeaconBlockHeaderNode) Slot() (Slot, error) {
	return SlotReadProp(PropReader(v, 0)).Slot()
}
func (v *BeaconBlockHeaderNode) SetSlot(s Slot) error {
	return SlotWriteProp(PropWriter(v, 0)).SetSlot(s)
}
func (v *BeaconBlockHeaderNode) ParentRoot() (Root, error) {
	return RootReadProp(PropReader(v, 1)).Root()
}
func (v *BeaconBlockHeaderNode) SetParentRoot(r Root) error {
	return RootWriteProp(PropWriter(v, 1)).SetRoot(r)
}
func (v *BeaconBlockHeaderNode) StateRoot() (Root, error) {
	return RootReadProp(PropReader(v, 2)).Root()
}
func (v *BeaconBlockHeaderNode) SetStateRoot(r Root) error {
	return RootWriteProp(PropWriter(v, 2)).SetRoot(r)
}
func (v *BeaconBlockHeaderNode) BodyRoot() (Root, error) {
	return RootReadProp(PropReader(v, 3)).Root()
}
func (v *BeaconBlockHeaderNode) SetBodyRoot(r Root) error {
	return RootWriteProp(PropWriter(v, 3)).SetRoot(r)
}

func (v *BeaconBlockHeaderNode) AsStruct() (*BeaconBlockHeader, error) {
	slot, err := v.Slot()
	if err != nil {
		return nil, err
	}
	parentRoot, err := v.ParentRoot()
	if err != nil {
		return nil, err
	}
	stateRoot, err := v.StateRoot()
	if err != nil {
		return nil, err
	}
	bodyRoot, err := v.BodyRoot()
	if err != nil {
		return nil, err
	}
	return &BeaconBlockHeader{
		Slot:       slot,
		ParentRoot: parentRoot,
		StateRoot:  stateRoot,
		BodyRoot:   bodyRoot,
	}, nil
}

type BeaconBlockHeaderReadProp ContainerReadProp

func (p BeaconBlockHeaderReadProp) BeaconBlockHeader() (*BeaconBlockHeaderNode, error) {
	if c, err := (ContainerReadProp)(p).Container(); err != nil {
		return nil, err
	} else {
		return &BeaconBlockHeaderNode{ContainerView: c}, nil
	}
}
