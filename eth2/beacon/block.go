package beacon

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zssz"
)

var BeaconBlockHeaderSSZ = zssz.GetSSZ((*BeaconBlockHeader)(nil))

type BeaconBlockHeader struct {
	Slot              Slot
	PreviousBlockRoot Root
	StateRoot         Root

	// Where the body would be, just a root embedded here.
	BlockBodyRoot Root
	// Signature
	Signature BLSSignature
}

var BeaconBlockBodySSZ = zssz.GetSSZ((*BeaconBlockBody)(nil))

type BeaconBlockBody struct {
	RandaoReveal BLSSignature
	Eth1Data     Eth1Data
	Graffiti     Root

	ProposerSlashings []ProposerSlashing
	AttesterSlashings []AttesterSlashing
	Attestations      []Attestation
	Deposits          []Deposit
	VoluntaryExits    []VoluntaryExit
	Transfers         []Transfer
}

var BeaconBlockSSZ = zssz.GetSSZ((*BeaconBlock)(nil))

type BeaconBlock struct {
	// Header
	Slot              Slot
	PreviousBlockRoot Root
	StateRoot         Root

	// Body
	Body BeaconBlockBody
	// Signature
	Signature BLSSignature
}
