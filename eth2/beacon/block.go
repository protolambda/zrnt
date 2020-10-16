package beacon

import (
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type SignedBeaconBlock struct {
	Message   BeaconBlock  `json:"message" yaml:"message"`
	Signature BLSSignature `json:"signature" yaml:"signature"`
}

func (b *SignedBeaconBlock) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Container(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedBeaconBlock) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Container(spec.Wrap(&b.Message), &b.Signature)
}

func (b *SignedBeaconBlock) ByteLength(spec *Spec) uint64 {
	return codec.ContainerLength(spec.Wrap(&b.Message), &b.Signature)
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
	Slot          Slot            `json:"slot" yaml:"slot"`
	ProposerIndex ValidatorIndex  `json:"proposer_index" yaml:"proposer_index"`
	ParentRoot    Root            `json:"parent_root" yaml:"parent_root"`
	StateRoot     Root            `json:"state_root" yaml:"state_root"`
	Body          BeaconBlockBody `json:"body" yaml:"body"`
}

func (b *BeaconBlock) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.Container(&b.Slot, &b.ProposerIndex, &b.ParentRoot, &b.StateRoot, spec.Wrap(&b.Body))
}

func (b *BeaconBlock) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.Container(&b.Slot, &b.ProposerIndex, &b.ParentRoot, &b.StateRoot, spec.Wrap(&b.Body))
}

func (b *BeaconBlock) ByteLength(spec *Spec) uint64 {
	return codec.ContainerLength(&b.Slot, &b.ProposerIndex, &b.ParentRoot, &b.StateRoot, spec.Wrap(&b.Body))
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
	RandaoReveal BLSSignature `json:"randao_reveal" yaml:"randao_reveal"`
	Eth1Data     Eth1Data     `json:"eth1_data" yaml:"eth1_data"`
	Graffiti     Root         `json:"graffiti" yaml:"graffiti"`

	ProposerSlashings ProposerSlashings `json:"proposer_slashings" yaml:"proposer_slashings"`
	AttesterSlashings AttesterSlashings `json:"attester_slashings" yaml:"attester_slashings"`
	Attestations      Attestations      `json:"attestations" yaml:"attestations"`
	Deposits          Deposits          `json:"deposits" yaml:"deposits"`
	VoluntaryExits    VoluntaryExits    `json:"voluntary_exits" yaml:"voluntary_exits"`
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
	return codec.ContainerLength(
		&b.RandaoReveal, &b.Eth1Data,
		&b.Graffiti, spec.Wrap(&b.ProposerSlashings),
		spec.Wrap(&b.AttesterSlashings), spec.Wrap(&b.Attestations),
		spec.Wrap(&b.Deposits), spec.Wrap(&b.VoluntaryExits),
	)
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
