package beacon

import (
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type SignedBeaconBlock struct {
	Message   BeaconBlock
	Signature BLSSignature
}

func (b *SignedBeaconBlock) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&b.Message, &b.Signature)
}

func (a *SignedBeaconBlock) FixedLength() uint64 {
	return 0
}

func (b *SignedBeaconBlock) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&b.Message, b.Signature)
}

func (block *SignedBeaconBlock) SignedHeader() *SignedBeaconBlockHeader {
	return &SignedBeaconBlockHeader{
		Message:   *block.Message.Header(),
		Signature: block.Signature,
	}
}

type BeaconBlock struct {
	Slot          Slot
	ProposerIndex ValidatorIndex
	ParentRoot    Root
	StateRoot     Root
	Body          BeaconBlockBody
}

func (b *BeaconBlock) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&b.Slot, &b.ProposerIndex, &b.ParentRoot, &b.StateRoot, &b.Body)
}

func (a *BeaconBlock) FixedLength() uint64 {
	return 0
}

func (b *BeaconBlock) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(b.Slot, b.ProposerIndex, b.ParentRoot, b.StateRoot, &b.Body)
}

func (c *Phase0Config) BeaconBlock() *ContainerTypeDef {
	return ContainerType("BeaconBlock", []FieldDef{
		{"slot", SlotType},
		{"proposer_index", ValidatorIndexType},
		{"parent_root", RootType},
		{"state_root", RootType},
		{"body", c.BeaconBlockBody()},
	})
}

func (c *Phase0Config) SignedBeaconBlock() *ContainerTypeDef {
	return ContainerType("SignedBeaconBlock", []FieldDef{
		{"message", c.BeaconBlock()},
		{"signature", BLSSignatureType},
	})
}

func (block *BeaconBlock) Header() *BeaconBlockHeader {
	return &BeaconBlockHeader{
		Slot:          block.Slot,
		ProposerIndex: block.ProposerIndex,
		ParentRoot:    block.ParentRoot,
		StateRoot:     block.StateRoot,
		BodyRoot:      block.Body.HashTreeRoot(tree.GetHashFn()),
	}
}

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

func (b *BeaconBlockBody) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(
		&b.RandaoReveal, &b.Eth1Data,
		&b.Graffiti, &b.ProposerSlashings,
		&b.AttesterSlashings, &b.Attestations,
		&b.Deposits, &b.VoluntaryExits,
	)
}

func (a *BeaconBlockBody) FixedLength() uint64 {
	return 0
}

func (b *BeaconBlockBody) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(
		b.RandaoReveal, &b.Eth1Data,
		b.Graffiti, &b.ProposerSlashings,
		&b.AttesterSlashings, &b.Attestations,
		&b.Deposits, &b.VoluntaryExits,
	)
}

func (b BeaconBlockBody) CheckLimits() error {
	if x := uint64(len(b.ProposerSlashings.Items)); x > b.ProposerSlashings.Limit {
		return fmt.Errorf("too many proposer slashings: %d", x)
	}
	if x := uint64(len(b.AttesterSlashings.Items)); x > b.AttesterSlashings.Limit {
		return fmt.Errorf("too many attester slashings: %d", x)
	}
	if x := uint64(len(b.Attestations.Items)); x > b.Attestations.Limit {
		return fmt.Errorf("too many attestations: %d", x)
	}
	if x := uint64(len(b.Deposits.Items)); x > b.Deposits.Limit {
		return fmt.Errorf("too many deposits: %d", x)
	}
	if x := uint64(len(b.VoluntaryExits.Items)); x > b.VoluntaryExits.Limit {
		return fmt.Errorf("too many voluntary exits: %d", x)
	}
	return nil
}

func (c *Phase0Config) BeaconBlockBody() *ContainerTypeDef {
	return ContainerType("BeaconBlockBody", []FieldDef{
		{"randao_reveal", BLSSignatureType},
		{"eth1_data", Eth1DataType}, // Eth1 data vote
		{"graffiti", Bytes32Type},   // Arbitrary data
		// Operations
		{"proposer_slashings", c.BlockProposerSlashings()},
		{"attester_slashings", c.BlockAttesterSlashings()},
		{"attestations", c.BlockAttestations()},
		{"deposits", c.BlockDeposits()},
		{"voluntary_exits", c.BlockVoluntaryExits()},
	})
}
