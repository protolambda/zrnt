package block_processing

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/beacon"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

func ProcessBlockHeader(state *BeaconState, block *BeaconBlock) error {
	// Verify that the slots match
	if block.Slot != state.Slot {
		return errors.New("slot of block does not match slot of state")
	}
	// Verify that the parent matches
	if signingRoot := ssz.SigningRoot(state.LatestBlockHeader, BeaconBlockHeaderSSZ); block.PreviousBlockRoot != signingRoot {
		return fmt.Errorf("previous block root %x does not match root %x from latest state block header", block.PreviousBlockRoot, signingRoot)
	}
	// Save current block as the new latest block
	state.LatestBlockHeader = BeaconBlockHeader{
		Slot:       block.Slot,
		ParentRoot: block.PreviousBlockRoot,
		BodyRoot:   ssz.HashTreeRoot(block.Body, BeaconBlockBodySSZ),
	}

	proposerIndex := state.GetBeaconProposerIndex()
	proposer := state.ValidatorRegistry[proposerIndex]
	// Verify proposer is not slashed
	if proposer.Slashed {
		return errors.New("cannot accept block header from slashed proposer")
	}
	// Block signature
	if !bls.BlsVerify(
		proposer.Pubkey,
		ssz.SigningRoot(block, BeaconBlockSSZ),
		block.Signature,
		state.GetDomain(DOMAIN_BEACON_PROPOSER, state.Epoch())) {
		return errors.New("block signature invalid")
	}
	return nil
}
