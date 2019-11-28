package header

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type Header interface {
	HashTreeRoot() Root
	Slot() (Slot, error)
	ParentRoot() (Root, error)
	BodyRoot() (Root, error)
}

var BeaconBlockHeaderType = &ContainerType{
	{"slot", SlotType},
	{"parent_root", RootType},
	{"state_root", RootType},
	{"body_root", RootType},
}

type BeaconBlockHeader struct { *ContainerView }

func NewBeaconBlockHeader() *BeaconBlockHeader {
	return &BeaconBlockHeader{ContainerView: BeaconBlockHeaderType.New()}
}

func (v *BeaconBlockHeader) HashTreeRoot() Root {
	return v.ViewRoot(tree.Hash)
}
func (v *BeaconBlockHeader) Slot() (Slot, error) {
	return SlotReadProp(PropReader(v, 0)).Slot()
}
func (v *BeaconBlockHeader) SetSlot(s Slot) error {
	return SlotWriteProp(PropWriter(v, 0)).SetSlot(s)
}
func (v *BeaconBlockHeader) ParentRoot() (Root, error) {
	return RootReadProp(PropReader(v, 1)).Root()
}
func (v *BeaconBlockHeader) SetParentRoot(r Root) error {
	return RootWriteProp(PropWriter(v, 1)).SetRoot(r)
}
func (v *BeaconBlockHeader) StateRoot() (Root, error) {
	return RootReadProp(PropReader(v, 2)).Root()
}
func (v *BeaconBlockHeader) SetStateRoot(r Root) error {
	return RootWriteProp(PropWriter(v, 2)).SetRoot(r)
}
func (v *BeaconBlockHeader) BodyRoot() (Root, error) {
	return RootReadProp(PropReader(v, 3)).Root()
}
func (v *BeaconBlockHeader) SetBodyRoot(r Root) error {
	return RootWriteProp(PropWriter(v, 3)).SetRoot(r)
}

type BeaconBlockHeaderReadProp ContainerReadProp

func (p BeaconBlockHeaderReadProp) BeaconBlockHeader() (*BeaconBlockHeader, error) {
	if c, err := (ContainerReadProp)(p).Container(); err != nil {
		return nil, err
	} else {
		return &BeaconBlockHeader{ContainerView: c}, nil
	}
}

type SignedBeaconBlockHeader struct { *ContainerView }

func (v *SignedBeaconBlockHeader) Message() (*BeaconBlockHeader, error) {
	return BeaconBlockHeaderReadProp(PropReader(v, 0)).BeaconBlockHeader()
}

func (v *SignedBeaconBlockHeader) Signature() (*BLSSignature, error) {
	return BLSSignatureReadProp(PropReader(v, 0)).BLSSignature()
}

type SignedBeaconBlockHeaderReadProp ContainerReadProp

func (p SignedBeaconBlockHeaderReadProp) SignedBeaconBlockHeader() (*SignedBeaconBlockHeader, error) {
	if c, err := (ContainerReadProp)(p).Container(); err != nil {
		return nil, err
	} else {
		return &SignedBeaconBlockHeader{ContainerView: c}, nil
	}
}
