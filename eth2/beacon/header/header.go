package header

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type HeaderProcessor interface {
	ProcessHeader(header *BeaconBlockHeader) error
}

type BlockHeaderFeature struct {
	State *BlockHeaderState
	Meta  interface {
		meta.Versioning
		meta.Proposers
		meta.Pubkeys
		meta.SlashedIndices
		meta.LatestHeader
		meta.LatestHeaderUpdate
	}
}

var BeaconBlockHeaderSSZ = zssz.GetSSZ((*BeaconBlockHeader)(nil))

type BeaconBlockHeader struct {
	Slot       Slot
	ParentRoot Root
	StateRoot  Root
	BodyRoot   Root // Where the body would be, just a root embedded here.
	Signature  BLSSignature
}

func (f *BlockHeaderFeature) ProcessHeader(header *BeaconBlockHeader) error {
	currentSlot := f.Meta.CurrentSlot()
	// Verify that the slots match
	if header.Slot != currentSlot {
		return errors.New("slot of block does not match slot of state")
	}
	// Verify that the parent matches
	if latestRoot := f.Meta.GetLatestBlockRoot(); header.ParentRoot != latestRoot {
		return fmt.Errorf("previous block root %x does not match root %x from latest state block header", header.ParentRoot, latestRoot)
	}

	proposerIndex := f.Meta.GetBeaconProposerIndex(currentSlot)
	// Verify proposer is not slashed
	if f.Meta.IsSlashed(proposerIndex) {
		return errors.New("cannot accept block header from slashed proposer")
	}
	// Block signature
	if !bls.BlsVerify(
		f.Meta.Pubkey(proposerIndex),
		ssz.SigningRoot(header, BeaconBlockHeaderSSZ),
		header.Signature,
		f.Meta.GetDomain(DOMAIN_BEACON_PROPOSER, f.Meta.CurrentEpoch())) {
		return errors.New("block signature invalid")
	}

	// Store as the new latest block
	f.State.LatestBlockHeader = BeaconBlockHeader{
		Slot:       header.Slot,
		ParentRoot: header.ParentRoot,
		// state_root is zeroed and overwritten in the next `process_slot` call.
		// with BlockHeaderState.UpdateStateRoot(), once the post state is available.
		BodyRoot: header.BodyRoot,
		// signature is always zeroed
	}

	return nil
}
