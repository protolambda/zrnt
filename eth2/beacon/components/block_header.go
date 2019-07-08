package components

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

var BeaconBlockHeaderSSZ = zssz.GetSSZ((*BeaconBlockHeader)(nil))

type BeaconBlockHeader struct {
	Slot       Slot
	ParentRoot Root
	StateRoot  Root
	// Where the body would be, just a root embedded here.
	BodyRoot Root
	// Signature
	Signature BLSSignature
}

func (header *BeaconBlockHeader) Process(state *BeaconState) error {
	// Verify that the slots match
	if header.Slot != state.Slot {
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
		BodyRoot:   header.BodyRoot,
		// note that StateRoot is set to 0. (filled by next process-slot call after block-processing)
	}

	proposerIndex := state.GetBeaconProposerIndex()
	proposer := state.Validators[proposerIndex]
	// Verify proposer is not slashed
	if proposer.Slashed {
		return errors.New("cannot accept block header from slashed proposer")
	}
	// Block signature
	if !bls.BlsVerify(
		proposer.Pubkey,
		ssz.SigningRoot(header, BeaconBlockHeaderSSZ),
		header.Signature,
		state.GetDomain(DOMAIN_BEACON_PROPOSER, state.Epoch())) {
		return errors.New("block signature invalid")
	}
	return nil
}
