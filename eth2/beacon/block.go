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

func (b *SignedBeaconBlock) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedBeaconBlock) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedBeaconBlock) ByteLength() uint64 {
	return 0 // TODO
}

func (a *SignedBeaconBlock) FixedLength(*Spec) uint64 {
	return 0
}

func (b *SignedBeaconBlock) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(spec.Wrap(&b.Message), b.Signature)
}

func (block *SignedBeaconBlock) SignedHeader(spec *Spec) *SignedBeaconBlockHeader {
	return &SignedBeaconBlockHeader{
		Message:   *block.Message.Header(spec),
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

func (b *BeaconBlock) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Container(&b.Slot, &b.ProposerIndex, &b.ParentRoot, &b.StateRoot, spec.Wrap(&b.Body))
}

func (b *BeaconBlock) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Container(&b.Slot, &b.ProposerIndex, &b.ParentRoot, &b.StateRoot, spec.Wrap(&b.Body))
}

func (b *BeaconBlock) ByteLength(spec *Spec) uint64 {
	return 0 // TODO
}

func (a *BeaconBlock) FixedLength(*Spec) uint64 {
	return 0
}

func (b *BeaconBlock) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(b.Slot, b.ProposerIndex, b.ParentRoot, b.StateRoot, spec.Wrap(&b.Body))
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

func (block *BeaconBlock) Header(spec *Spec) *BeaconBlockHeader {
	return &BeaconBlockHeader{
		Slot:          block.Slot,
		ProposerIndex: block.ProposerIndex,
		ParentRoot:    block.ParentRoot,
		StateRoot:     block.StateRoot,
		BodyRoot:      block.Body.HashTreeRoot(spec, tree.GetHashFn()),
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

func (b *BeaconBlockBody) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Container(
		&b.RandaoReveal, &b.Eth1Data,
		&b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
	)
}

func (b *BeaconBlockBody) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Container(
		&b.RandaoReveal, &b.Eth1Data,
		&b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
	)
}

func (b *BeaconBlockBody) ByteLength(spec *Spec) uint64 {
	return 0 // TODO
}

func (a *BeaconBlockBody) FixedLength(*Spec) uint64 {
	return 0
}

func (b *BeaconBlockBody) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(
		b.RandaoReveal, &b.Eth1Data,
		b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
	)
}

func (b BeaconBlockBody) CheckLimits(spec *Spec) error {
	if x := uint64(len(b.ProposerSlashings)); x > spec.MAX_PROPOSER_SLASHINGS {
		return fmt.Errorf("too many proposer slashings: %d", x)
	}
	if x := uint64(len(b.AttesterSlashings)); x > spec.MAX_ATTESTER_SLASHINGS {
		return fmt.Errorf("too many attester slashings: %d", x)
	}
	if x := uint64(len(b.Attestations)); x > spec.MAX_ATTESTATIONS {
		return fmt.Errorf("too many attestations: %d", x)
	}
	if x := uint64(len(b.Deposits)); x > spec.MAX_DEPOSITS {
		return fmt.Errorf("too many deposits: %d", x)
	}
	if x := uint64(len(b.VoluntaryExits)); x > spec.MAX_VOLUNTARY_EXITS {
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
