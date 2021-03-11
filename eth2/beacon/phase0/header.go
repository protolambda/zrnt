package phase0

import (
	"context"
	"errors"
	"fmt"

	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type BeaconBlockHeader struct {
	Slot          common.Slot           `json:"slot" yaml:"slot"`
	ProposerIndex common.ValidatorIndex `json:"proposer_index" yaml:"proposer_index"`
	ParentRoot    common.Root           `json:"parent_root" yaml:"parent_root"`
	StateRoot     common.Root           `json:"state_root" yaml:"state_root"`
	BodyRoot      common.Root           `json:"body_root" yaml:"body_root"`
}

func (h *BeaconBlockHeader) View() *BeaconBlockHeaderView {
	pr := RootView(h.ParentRoot)
	sr := RootView(h.StateRoot)
	br := RootView(h.BodyRoot)
	c, _ := BeaconBlockHeaderType.FromFields(
		Uint64View(h.Slot),
		Uint64View(h.ProposerIndex),
		&pr,
		&sr,
		&br,
	)
	return &BeaconBlockHeaderView{c}
}

func (s *BeaconBlockHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&s.Slot, &s.ProposerIndex, &s.ParentRoot, &s.StateRoot, &s.BodyRoot)
}

func (s *BeaconBlockHeader) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&s.Slot, &s.ProposerIndex, &s.ParentRoot, &s.StateRoot, &s.BodyRoot)
}

func (s *BeaconBlockHeader) ByteLength() uint64 {
	return BeaconBlockHeaderType.TypeByteLength()
}

func (b *BeaconBlockHeader) FixedLength() uint64 {
	return BeaconBlockHeaderType.TypeByteLength()
}

func (b *BeaconBlockHeader) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(b.Slot, b.ProposerIndex, b.ParentRoot, b.StateRoot, b.BodyRoot)
}

type SignedBeaconBlockHeader struct {
	Message   BeaconBlockHeader   `json:"message" yaml:"message"`
	Signature common.BLSSignature `json:"signature" yaml:"signature"`
}

func (s *SignedBeaconBlockHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&s.Message, &s.Signature)
}

func (s *SignedBeaconBlockHeader) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&s.Message, &s.Signature)
}

func (s *SignedBeaconBlockHeader) ByteLength() uint64 {
	return SignedBeaconBlockHeaderType.TypeByteLength()
}

func (b *SignedBeaconBlockHeader) FixedLength() uint64 {
	return SignedBeaconBlockHeaderType.TypeByteLength()
}

func (s *SignedBeaconBlockHeader) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&s.Message, s.Signature)
}

var SignedBeaconBlockHeaderType = ContainerType("SignedBeaconBlockHeader", []FieldDef{
	{"message", BeaconBlockHeaderType},
	{"signature", common.BLSSignatureType},
})

var BeaconBlockHeaderType = ContainerType("BeaconBlockHeader", []FieldDef{
	{"slot", common.SlotType},
	{"proposer_index", common.ValidatorIndexType},
	{"parent_root", RootType},
	{"state_root", RootType},
	{"body_root", RootType},
})

type BeaconBlockHeaderView struct {
	*ContainerView
}

func AsBeaconBlockHeader(v View, err error) (*BeaconBlockHeaderView, error) {
	c, err := AsContainer(v, err)
	return &BeaconBlockHeaderView{c}, err
}

func (v *BeaconBlockHeaderView) Slot() (common.Slot, error) {
	return common.AsSlot(v.Get(0))
}

func (v *BeaconBlockHeaderView) ProposerIndex() (common.ValidatorIndex, error) {
	return common.AsValidatorIndex(v.Get(1))
}

func (v *BeaconBlockHeaderView) ParentRoot() (common.Root, error) {
	return AsRoot(v.Get(2))
}

func (v *BeaconBlockHeaderView) StateRoot() (common.Root, error) {
	return AsRoot(v.Get(3))
}

func (v *BeaconBlockHeaderView) SetStateRoot(root common.Root) error {
	rv := RootView(root)
	return v.Set(3, &rv)
}

func (v *BeaconBlockHeaderView) BodyRoot() (common.Root, error) {
	return AsRoot(v.Get(4))
}

func (v *BeaconBlockHeaderView) Raw() (*BeaconBlockHeader, error) {
	slot, err := v.Slot()
	if err != nil {
		return nil, err
	}
	parentRoot, err := v.ParentRoot()
	if err != nil {
		return nil, err
	}
	stateRoot, err := v.StateRoot()
	if err != nil {
		return nil, err
	}
	bodyRoot, err := v.BodyRoot()
	if err != nil {
		return nil, err
	}
	return &BeaconBlockHeader{
		Slot:       slot,
		ParentRoot: parentRoot,
		StateRoot:  stateRoot,
		BodyRoot:   bodyRoot,
	}, nil
}

func ProcessHeader(ctx context.Context, spec *common.Spec, epc *EpochsContext, state *BeaconStateView, header *BeaconBlock) error {
	select {
	case <-ctx.Done():
		return common.TransitionCancelErr
	default: // Don't block.
		break
	}
	currentSlot, err := state.Slot()
	if err != nil {
		return err
	}
	// Verify that the slots match
	if header.Slot != currentSlot {
		return errors.New("slot of block does not match slot of state")
	}
	latestHeader, err := state.LatestBlockHeader()
	if err != nil {
		return err
	}
	latestSlot, err := latestHeader.Slot()
	if err != nil {
		return err
	}
	if header.Slot <= latestSlot {
		return errors.New("bad block header")
	}
	if isValid, err := state.IsValidIndex(header.ProposerIndex); err != nil {
		return err
	} else if !isValid {
		return fmt.Errorf("beacon block header proposer index is out of range: %d", header.ProposerIndex)
	}
	proposerIndex, err := epc.GetBeaconProposer(currentSlot)
	if err != nil {
		return err
	}
	if header.ProposerIndex != proposerIndex {
		return fmt.Errorf("beacon block header proposer index does not match expected index: got: %d, expected: %d", header.ProposerIndex, proposerIndex)
	}
	// Verify that the parent matches
	latestRoot := latestHeader.HashTreeRoot(tree.GetHashFn())
	if header.ParentRoot != latestRoot {
		return fmt.Errorf("previous block root %x does not match root %x from latest state block header", header.ParentRoot, latestRoot)
	}
	validators, err := state.Validators()
	if err != nil {
		return err
	}
	validator, err := validators.Validator(proposerIndex)
	if err != nil {
		return err
	}
	// Verify proposer is not slashed
	if slashed, err := validator.Slashed(); err != nil {
		return err
	} else if slashed {
		return errors.New("cannot accept block header from slashed proposer")
	}

	// Store as the new latest block
	headerRaw := BeaconBlockHeader{
		Slot:          header.Slot,
		ProposerIndex: header.ProposerIndex,
		ParentRoot:    header.ParentRoot,
		// state_root is zeroed and overwritten in the next `process_slot` call.
		// with BlockHeaderState.UpdateStateRoot(), once the post state is available.
		StateRoot: common.Root{},
		BodyRoot:  header.Body.HashTreeRoot(spec, tree.GetHashFn()),
	}
	return state.SetLatestBlockHeader(headerRaw.View())
}
