package beacon

import "go-beacon-transition/eth2"

type BeaconBlock struct {
	// Header
	Slot          eth2.Slot
	Parent_root   eth2.Root
	State_root    eth2.Root
	Randao_reveal [96]byte
	Eth1_data     Eth1Data

	// Body
	Body BeaconBlockBody
	// Signature
	Signature eth2.BLSSignature
}

type BeaconBlockBody struct {
	Proposer_slashings []ProposerSlashing
	Attester_slashings []AttesterSlashing
	Attestations       []Attestation
	Deposits           []Deposit
	Voluntary_exits    []VoluntaryExit
	Transfers          []Transfer
}

