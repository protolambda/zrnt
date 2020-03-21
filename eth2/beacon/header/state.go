package header

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
)

type HeaderProcessor interface {
	ProcessHeader(input BlockHeaderProcessInput, header *BeaconBlockHeader) error
}

type BlockHeaderProcessInput interface {
	meta.Versioning
	meta.Proposers
	meta.Pubkeys
	meta.SlashedIndices
	meta.LatestHeader
	meta.LatestHeaderUpdate
}

type LatestBlockHeaderProp BeaconBlockHeaderReadProp

func (p LatestBlockHeaderProp) GetLatestHeader() (*BeaconBlockHeaderNode, error) {
	return BeaconBlockHeaderReadProp(p).BeaconBlockHeader()
}

func (p LatestBlockHeaderProp) GetLatestBlockRoot() (Root, error) {
	h, err := p.GetLatestHeader()
	if err != nil {
		return Root{}, err
	}
	return h.HashTreeRoot(), nil
}

func (p LatestBlockHeaderProp) UpdateLatestBlockStateRoot(root Root) error {
	prev, err := BeaconBlockHeaderReadProp(p).BeaconBlockHeader()
	if err != nil {
		return err
	}
	return prev.SetStateRoot(root)
}

func ProcessHeader(input BlockHeaderProcessInput, header *BeaconBlockHeader) error {
	currentSlot, err := input.CurrentSlot()
	if err != nil {
		return err
	}
	// Verify that the slots match
	if header.Slot != currentSlot {
		return errors.New("slot of block does not match slot of state")
	}
	// Verify that the parent matches
	if latestRoot, err := input.GetLatestBlockRoot(); err != nil {
		return err
	} else if header.ParentRoot != latestRoot {
		return fmt.Errorf("previous block root %x does not match root %x from latest state block header", header.ParentRoot, latestRoot)
	}

	proposerIndex, err := input.GetBeaconProposerIndex(currentSlot)
	if err != nil {
		return err
	}
	// Verify proposer is not slashed
	if slashed, err := input.IsSlashed(proposerIndex); err != nil {
		return err
	} else if slashed {
		return errors.New("cannot accept block header from slashed proposer")
	}

	newLatest := NewBeaconBlockHeaderNode()
	if err := newLatest.SetSlot(header.Slot); err != nil {
		return err
	}
	if err := newLatest.SetParentRoot(header.ParentRoot); err != nil {
		return err
	}
	if err := newLatest.SetBodyRoot(header.StateRoot); err != nil {
		return err
	}
	// state_root is zeroed and overwritten in the next `process_slot` call.
	// with BlockHeaderState.UpdateStateRoot(), once the post state is available.

	// Store as the new latest block
	h, err := f.State.GetLatestHeader()
	if err != nil {
		return err
	}
	return h.PropagateChange(newLatest)
}
