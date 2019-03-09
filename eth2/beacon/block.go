package beacon

import (
	"github.com/protolambda/zrnt/eth2/util/ssz"
)

type BeaconBlockHeader struct {
	Slot              Slot
	PreviousBlockRoot Root
	StateRoot         Root

	// Where the body would be, just a root embedded here.
	BlockBodyRoot Root
	// Signature
	Signature BLSSignature `ssz:"signature"`
}

type BeaconBlockBody struct {
	RandaoReveal BLSSignature
	Eth1Data     Eth1Data

	ProposerSlashings []ProposerSlashing
	AttesterSlashings []AttesterSlashing
	Attestations      []Attestation
	Deposits          []Deposit
	VoluntaryExits    []VoluntaryExit
	Transfers         []Transfer
}

type BeaconBlock struct {
	// Header
	Slot              Slot
	PreviousBlockRoot Root
	StateRoot         Root

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
		StateRoot:         Root{}, // empty hash, "temporary" part.
		BlockBodyRoot:     ssz.HashTreeRoot(block.Body),
		Signature:         block.Signature,
	}
}
