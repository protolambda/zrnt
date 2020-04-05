package beacon

import (
	. "github.com/protolambda/zrnt/eth2/beacon/attestations"
	. "github.com/protolambda/zrnt/eth2/beacon/deposits"
	. "github.com/protolambda/zrnt/eth2/beacon/eth1"
	. "github.com/protolambda/zrnt/eth2/beacon/exits"
	. "github.com/protolambda/zrnt/eth2/beacon/header"
	. "github.com/protolambda/zrnt/eth2/beacon/randao"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/attslash"
	. "github.com/protolambda/zrnt/eth2/beacon/slashings/propslash"

	"github.com/protolambda/zrnt/eth2/util/bls"
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

var BeaconBlockType = &ContainerType{
	{"slot", SlotType},
	{"proposer_index", ValidatorIndexType},
	{"parent_root", RootType},
	{"state_root", RootType},
	{"body", BeaconBlockBodyType},
}

var SignedBeaconBlockType = &ContainerType{
	{"message", BeaconBlockType},
	{"signature", BLSSignatureType},
}

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

var BeaconBlockBodyType = &ContainerType{
	{"randao_reveal", BLSSignatureType},
	{"eth1_data", Eth1DataType}, // Eth1 data vote
	{"graffiti", Bytes32Type},   // Arbitrary data
	// Operations
	{"proposer_slashings", ProposerSlashingsType},
	{"attester_slashings", AttesterSlashingsType},
	{"attestations", AttestationsType},
	{"deposits", DepositsType},
	{"voluntary_exits", VoluntaryExitsType},
}

func (f *BlockProcessFeature) Slot() Slot {
	return f.Block.Message.Slot
}

func (f *BlockProcessFeature) StateRoot() Root {
	return f.Block.Message.StateRoot
}

func (f *BlockProcessFeature) VerifyStateRoot(expected Root) bool {
	return f.Block.Message.StateRoot == expected
}

func (f *BlockProcessFeature) BlockRoot() Root {
	return ssz.HashTreeRoot(&f.Block.Message, BeaconBlockSSZ)
}

func (f *BlockProcessFeature) Signature() BLSSignature {
	return f.Block.Signature
}

func (f *BlockProcessFeature) VerifySignature(pubkey BLSPubkey, version Version, genValRoot Root) bool {
	return bls.Verify(
		pubkey,
		ComputeSigningRoot(
			f.BlockRoot(),
			ComputeDomain(DOMAIN_BEACON_PROPOSER, version, genValRoot)),
		f.Signature())
}

func (f *BlockProcessFeature) Process() error {
	header := f.Block.Message.Header()
	if err := f.Meta.ProcessHeader(header); err != nil {
		return err
	}
	body := &f.Block.Message.Body
	if err := f.Meta.ProcessRandaoReveal(body.RandaoReveal); err != nil {
		return err
	}
	if err := f.Meta.ProcessEth1Vote(body.Eth1Data); err != nil {
		return err
	}
	if err := f.Meta.ProcessProposerSlashings(body.ProposerSlashings); err != nil {
		return err
	}
	if err := f.Meta.ProcessAttesterSlashings(body.AttesterSlashings); err != nil {
		return err
	}
	if err := f.Meta.ProcessAttestations(body.Attestations); err != nil {
		return err
	}
	if err := f.Meta.ProcessDeposits(body.Deposits); err != nil {
		return err
	}
	if err := f.Meta.ProcessVoluntaryExits(body.VoluntaryExits); err != nil {
		return err
	}
	return nil
}
