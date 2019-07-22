package header

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

type BlockHeaderFeature struct {
	*BlockHeaderState
	Meta interface {
		VersioningMeta
		ProposingMeta
		PubkeyMeta
		SlashedMeta
		HeaderMeta
		UpdateHeaderMeta
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

func (state *BlockHeaderFeature) ProcessHeader(header *BeaconBlockHeader) error {
	currentSlot := state.Meta.CurrentSlot()
	// Verify that the slots match
	if header.Slot != currentSlot {
		return errors.New("slot of block does not match slot of state")
	}
	// Verify that the parent matches
	if latestRoot := state.Meta.GetLatestBlockRoot(); header.ParentRoot != latestRoot {
		return fmt.Errorf("previous block root %x does not match root %x from latest state block header", header.ParentRoot, latestRoot)
	}

	proposerIndex := state.Meta.GetBeaconProposerIndex(currentSlot)
	// Verify proposer is not slashed
	if state.Meta.IsSlashed(proposerIndex) {
		return errors.New("cannot accept block header from slashed proposer")
	}
	// Block signature
	if !bls.BlsVerify(
		state.Meta.Pubkey(proposerIndex),
		ssz.SigningRoot(header, BeaconBlockHeaderSSZ),
		header.Signature,
		state.Meta.GetDomain(DOMAIN_BEACON_PROPOSER, state.Meta.CurrentEpoch())) {
		return errors.New("block signature invalid")
	}

	// Store as the new latest block
	state.LatestBlockHeader = BeaconBlockHeader{
		Slot:       header.Slot,
		ParentRoot: header.ParentRoot,
		// state_root: zeroed, overwritten with UpdateStateRoot(), once the post state is available.
		BodyRoot: header.BodyRoot,
		// signature is always zeroed
	}

	return nil
}
