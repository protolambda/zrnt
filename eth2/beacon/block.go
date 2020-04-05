package beacon

import (
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
	. "github.com/protolambda/ztyp/view"
)

var SignedBeaconBlockSSZ = zssz.GetSSZ((*SignedBeaconBlock)(nil))

type SignedBeaconBlock struct {
	Message   BeaconBlock
	Signature BLSSignature
}

func (block *SignedBeaconBlock) SignedHeader() *SignedBeaconBlockHeader {
	return &SignedBeaconBlockHeader{
		Message:   *block.Message.Header(),
		Signature: block.Signature,
	}
}

var BeaconBlockSSZ = zssz.GetSSZ((*BeaconBlock)(nil))

type BeaconBlock struct {
	Slot          Slot
	ProposerIndex ValidatorIndex
	ParentRoot    Root
	StateRoot     Root
	Body          BeaconBlockBody
}

var BeaconBlockType =  ContainerType("BeaconBlock", []FieldDef{
	{"slot", SlotType},
	{"proposer_index", ValidatorIndexType},
	{"parent_root", RootType},
	{"state_root", RootType},
	{"body", BeaconBlockBodyType},
})

var SignedBeaconBlockType = ContainerType("SignedBeaconBlock", []FieldDef{
	{"message", BeaconBlockType},
	{"signature", BLSSignatureType},
})

func (block *BeaconBlock) Header() *BeaconBlockHeader {
	return &BeaconBlockHeader{
		Slot:          block.Slot,
		ProposerIndex: block.ProposerIndex,
		ParentRoot:    block.ParentRoot,
		StateRoot:     block.StateRoot,
		BodyRoot:      block.Body.HashTreeRoot(),
	}
}

var BeaconBlockBodySSZ = zssz.GetSSZ((*BeaconBlockBody)(nil))

type BeaconBlockBody struct {
	RandaoReveal BLSSignature
	Eth1Data     Eth1Data // Eth1 data vote
	Graffiti     Root     // Arbitrary data

	ProposerSlashings ProposerSlashings
	AttesterSlashings AttesterSlashings
	Attestations      Attestations
	Deposits          Deposits
	VoluntaryExits    VoluntaryExits
}

func (b *BeaconBlockBody) HashTreeRoot() Root {
	return ssz.HashTreeRoot(b, BeaconBlockBodySSZ)
}

var BeaconBlockBodyType = ContainerType("BeaconBlockBody", []FieldDef{
	{"randao_reveal", BLSSignatureType},
	{"eth1_data", Eth1DataType}, // Eth1 data vote
	{"graffiti", Bytes32Type},   // Arbitrary data
	// Operations
	{"proposer_slashings", ProposerSlashingsType},
	{"attester_slashings", AttesterSlashingsType},
	{"attestations", AttestationsType},
	{"deposits", DepositsType},
	{"voluntary_exits", VoluntaryExitsType},
})
