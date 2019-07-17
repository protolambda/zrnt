package header

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon/components/meta"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type BlockHeaderState struct {
	LatestBlockHeader BeaconBlockHeader
}

type BlockHeaderReq interface {
	VersioningMeta
	ProposingMeta
	CompactValidatorMeta
}

func (state *BlockHeaderState) ProcessBlockHeader(meta BlockHeaderReq, header *BeaconBlockHeader) error {
	// Verify that the slots match
	if header.Slot != meta.Slot() {
		return errors.New("slot of block does not match slot of state")
	}
	// Verify that the parent matches
	if signingRoot := ssz.SigningRoot(state.LatestBlockHeader, BeaconBlockHeaderSSZ); header.ParentRoot != signingRoot {
		return fmt.Errorf("previous block root %x does not match root %x from latest state block header", header.ParentRoot, signingRoot)
	}
	// Save current block as the new latest block
	state.LatestBlockHeader = BeaconBlockHeader{
		Slot:       header.Slot,
		ParentRoot: header.ParentRoot,
		// state_root: zeroed, overwritten in the next ProcessSlot call
		BodyRoot: header.BodyRoot,
		// signature is always zeroed
	}

	proposerIndex := meta.GetBeaconProposerIndex()
	// Verify proposer is not slashed
	if meta.IsSlashed(proposerIndex) {
		return errors.New("cannot accept block header from slashed proposer")
	}
	// Block signature
	if !bls.BlsVerify(
		meta.Pubkey(proposerIndex),
		ssz.SigningRoot(header, BeaconBlockHeaderSSZ),
		header.Signature,
		meta.GetDomain(DOMAIN_BEACON_PROPOSER, meta.Epoch())) {
		return errors.New("block signature invalid")
	}
	return nil
}
