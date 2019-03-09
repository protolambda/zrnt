package beacon

import (
	"github.com/protolambda/go-beacon-transition/eth2/util/ssz"
)

type BeaconBlockHeader struct {
	Slot              Slot
	PreviousBlockRoot Root
	State_root        Root

	// Where the body would be, just a root embedded here.
	BlockBodyRoot Root
	// Signature
	Signature BLSSignature `ssz:"signature"`
}

type BeaconBlockBody struct {
	Randao_reveal BLSSignature
	Eth1_data     Eth1Data

	Proposer_slashings []ProposerSlashing
	Attester_slashings []AttesterSlashing
	Attestations       []Attestation
	Deposits           []Deposit
	Voluntary_exits    []VoluntaryExit
	Transfers          []Transfer
}


type BeaconBlock struct {
	// Header
	Slot              Slot
	PreviousBlockRoot Root
	State_root        Root

	// Body
	Body BeaconBlockBody
	// Signature
	Signature BLSSignature `ssz:"signature"`
}

// Returns a template for a genesis block
// (really just all default 0 data, but with genesis slot initialized)
func GetEmptyBlock() *BeaconBlock {
	return &BeaconBlock{
		Slot: GENESIS_SLOT,
	}
}

// Return the block header corresponding to a block with ``state_root`` set to ``ZERO_HASH``.
func (block *BeaconBlock) GetTemporaryBlockHeader() BeaconBlockHeader {
	return BeaconBlockHeader{
		Slot:              block.Slot,
		PreviousBlockRoot: block.PreviousBlockRoot,
		State_root:        Root{}, // empty hash, "temporary" part.
		BlockBodyRoot:     ssz.Hash_tree_root(block.Body),
		Signature:         block.Signature,
	}
}
