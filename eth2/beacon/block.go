package beacon

import "github.com/protolambda/go-beacon-transition/eth2"

type BeaconBlock struct {
	// Header
	Slot              eth2.Slot
	PreviousBlockRoot eth2.Root
	State_root        eth2.Root

	//// Body
	Body BeaconBlockBody
	//// Signature
	Signature eth2.BLSSignature `ssz:"signature"`
}

// Returns a template for a genesis block
// (really just all default 0 data, but with genesis slot initialized)
func GetEmptyBlock() *BeaconBlock {
	return &BeaconBlock{
		Slot: eth2.GENESIS_SLOT,
	}
}

// Return the block header corresponding to a block with ``state_root`` set to ``ZERO_HASH``.
func (block *BeaconBlock) GetTemporaryBlockHeader() *BeaconBlockHeader {
	return &BeaconBlockHeader{
		Slot: block.Slot,
		PreviousBlockRoot: block.PreviousBlockRoot,
		State_root: block.State_root,
		BlockBodyRoot: eth2.Root{},// empty hash, "temporary" part.
		Signature: block.Signature,
	}
}

type BeaconBlockHeader struct {
	Slot eth2.Slot
	PreviousBlockRoot eth2.Root
	State_root        eth2.Root

	// Where the body would be, just a root embedded here.
	BlockBodyRoot eth2.Root
	// Signature
	Signature eth2.BLSSignature `ssz:"signature"`
}

type BeaconBlockBody struct {
	Randao_reveal [96]byte
	Eth1_data     Eth1Data

	Proposer_slashings []ProposerSlashing
	Attester_slashings []AttesterSlashing
	Attestations       []Attestation
	Deposits           []Deposit
	Voluntary_exits    []VoluntaryExit
	Transfers          []Transfer
}

