package header

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	. "github.com/protolambda/ztyp/props"
)

type HeaderProcessor interface {
	ProcessHeader(header Header) error
}

type BlockHeaderFeature struct {
	State *LatestHeaderWriteProp
	Meta  interface {
		meta.Versioning
		meta.Proposers
		meta.Pubkeys
		meta.SlashedIndices
		meta.LatestHeader
		meta.LatestHeaderUpdate
	}
}

type LatestHeaderWriteProp WritePropFn

func (p LatestHeaderWriteProp) SetLatestHeader(v *BeaconBlockHeader) error {
	return p(v)
}

type LatestHeaderProp struct {
	BeaconBlockHeaderReadProp
	LatestHeaderWriteProp
}

func (p LatestHeaderProp) UpdateStateRoot(root Root) error {
	prev, err := p.BeaconBlockHeaderReadProp.BeaconBlockHeader()
	if err != nil {
		return err
	}
	// modifying the view will only fork the view from the original tree, i.e. a copy.
	if err := prev.SetStateRoot(root); err != nil {
		return err
	}
	return p.LatestHeaderWriteProp.SetLatestHeader(prev)
}

func (f *BlockHeaderFeature) ProcessHeader(header Header) error {
	currentSlot, err := f.Meta.CurrentSlot()
	if err != nil {
		return err
	}
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
	if latestRoot, err := f.Meta.GetLatestBlockRoot(); err != nil {
		return err
	} else if parentRoot != latestRoot {
		return fmt.Errorf("previous block root %x does not match root %x from latest state block header", parentRoot, latestRoot)
	}

	proposerIndex, err := f.Meta.GetBeaconProposerIndex(currentSlot)
	if err != nil {
		return err
	}
	// Verify proposer is not slashed
	if slashed, err := f.Meta.IsSlashed(proposerIndex); err != nil {
		return err
	} else if slashed {
		return errors.New("cannot accept block header from slashed proposer")
	}

	newLatest := NewBeaconBlockHeader()
	if err := newLatest.SetSlot(headerSlot); err != nil {
		return err
	}
	if err := newLatest.SetParentRoot(parentRoot); err != nil {
		return err
	}
	if err := newLatest.SetBodyRoot(bodyRoot); err != nil {
		return err
	}
	// state_root is zeroed and overwritten in the next `process_slot` call.
	// with BlockHeaderState.UpdateStateRoot(), once the post state is available.

	// Store as the new latest block
	return f.State.SetLatestHeader(newLatest)
}
