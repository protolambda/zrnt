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

type BlockHeaderReq interface {
	VersioningMeta
	ProposingMeta
	CompactValidatorMeta
	HeaderMeta
	UpdateHeaderMeta
}

var BeaconBlockHeaderSSZ = zssz.GetSSZ((*BeaconBlockHeader)(nil))

type BeaconBlockHeader struct {
	Slot       Slot
	ParentRoot Root
	StateRoot  Root
	BodyRoot   Root // Where the body would be, just a root embedded here.
	Signature  BLSSignature
}

func (header *BeaconBlockHeader) Process(meta BlockHeaderReq) error {
	// Verify that the slots match
	if header.Slot != meta.Slot() {
		return errors.New("slot of block does not match slot of state")
	}
	// Verify that the parent matches
	if latestRoot := meta.GetLatestBlockRoot(); header.ParentRoot != latestRoot {
		return fmt.Errorf("previous block root %x does not match root %x from latest state block header", header.ParentRoot, latestRoot)
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

	// Store as the new latest block
	meta.StoreHeaderData(header.Slot, header.ParentRoot, header.BodyRoot)

	return nil
}
