package header

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

type Header interface {
	HashTreeRoot() (Root, error)
	Slot() (Slot, error)
	ParentRoot() (Root, error)
	BodyRoot() (Root, error)
}

type HeaderProcessor interface {
	ProcessHeader(header Header) error
}

type BlockHeaderFeature struct {
	Meta  interface {
		SetLatestHeader(header *BeaconBlockHeader)
		meta.Versioning
		meta.Proposers
		meta.Pubkeys
		meta.SlashedIndices
		meta.LatestHeader
		meta.LatestHeaderUpdate
	}
}

var BeaconBlockHeaderType = &ContainerType{
	{"slot", SlotType},
	{"parent_root", RootType},
	{"state_root", RootType},
	{"body_root", RootType},
	{"signature", BLSSignatureType},
}

type BeaconBlockHeader struct { *ContainerView }

func NewBeaconBlockHeader() *BeaconBlockHeader {
	return &BeaconBlockHeader{ContainerView: BeaconBlockHeaderType.New()}
}


func (v *BeaconBlockHeader) HashTreeRoot() (Root, error) {

}
func (v *BeaconBlockHeader) Slot() (Slot, error) {

}
func (v *BeaconBlockHeader) ParentRoot() (Root, error) {

}
func (v *BeaconBlockHeader) BodyRoot() (Root, error) {

}

func (f *BlockHeaderFeature) ProcessHeader(header Header) error {
	currentSlot := f.Meta.CurrentSlot()
	headerSlot, err := header.Slot()
	if err != nil {
		return err
	}
	parentRoot, err := header.ParentRoot()
	if err != nil {
		return err
	}
	bodyRoot, err := header.BodyRoot()
	if err != nil {
		return err
	}
	// Verify that the slots match
	if headerSlot != currentSlot {
		return errors.New("slot of block does not match slot of state")
	}
	// Verify that the parent matches
	if latestRoot := f.Meta.GetLatestBlockRoot(); parentRoot != latestRoot {
		return fmt.Errorf("previous block root %x does not match root %x from latest state block header", parentRoot, latestRoot)
	}

	proposerIndex := f.Meta.GetBeaconProposerIndex(currentSlot)
	// Verify proposer is not slashed
	if slashed, err := f.Meta.IsSlashed(proposerIndex); err != nil {
		return err
	} else if slashed {
		return errors.New("cannot accept block header from slashed proposer")
	}

	// Store as the new latest block
	f.State.LatestBlockHeader = BeaconBlockHeader{
		Slot:       headerSlot,
		ParentRoot: parentRoot,
		// state_root is zeroed and overwritten in the next `process_slot` call.
		// with BlockHeaderState.UpdateStateRoot(), once the post state is available.
		BodyRoot: bodyRoot,
		// signature is always zeroed
	}
	return nil
}
