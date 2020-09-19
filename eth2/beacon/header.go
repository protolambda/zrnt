package beacon

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type BeaconBlockHeader struct {
	Slot          Slot
	ProposerIndex ValidatorIndex
	ParentRoot    Root
	StateRoot     Root
	BodyRoot      Root
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
	return dr.Container(&s.Slot, &s.ProposerIndex, &s.ParentRoot, &s.StateRoot, &s.BodyRoot)
}

func (b *BeaconBlockHeader) FixedLength() uint64 {
	return BeaconBlockHeaderType.TypeByteLength()
}

func (b *BeaconBlockHeader) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(b.Slot, b.ProposerIndex, b.ParentRoot, b.StateRoot, b.BodyRoot)
}

type SignedBeaconBlockHeader struct {
	Message   BeaconBlockHeader
	Signature BLSSignature
}

func (s *SignedBeaconBlockHeader) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&s.Message, &s.Signature)
}

func (b *SignedBeaconBlockHeader) FixedLength() uint64 {
	return SignedBeaconBlockHeaderType.TypeByteLength()
}

func (s *SignedBeaconBlockHeader) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&s.Message, s.Signature)
}

var SignedBeaconBlockHeaderType = ContainerType("SignedBeaconBlockHeader", []FieldDef{
	{"message", BeaconBlockHeaderType},
	{"signature", BLSSignatureType},
})

var BeaconBlockHeaderType = ContainerType("BeaconBlockHeader", []FieldDef{
	{"slot", SlotType},
	{"proposer_index", ValidatorIndexType},
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

func (v *BeaconBlockHeaderView) Slot() (Slot, error) {
	return AsSlot(v.Get(0))
}

func (v *BeaconBlockHeaderView) ProposerIndex() (ValidatorIndex, error) {
	return AsValidatorIndex(v.Get(1))
}

func (v *BeaconBlockHeaderView) ParentRoot() (Root, error) {
	return AsRoot(v.Get(2))
}

func (v *BeaconBlockHeaderView) StateRoot() (Root, error) {
	return AsRoot(v.Get(3))
}

func (v *BeaconBlockHeaderView) SetStateRoot(root Root) error {
	rv := RootView(root)
	return v.Set(3, &rv)
}

func (v *BeaconBlockHeaderView) BodyRoot() (Root, error) {
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

func (spec *Spec) ProcessHeader(ctx context.Context, epc *EpochsContext, state *BeaconStateView, header *BeaconBlock) error {
	select {
	case <-ctx.Done():
		return TransitionCancelErr
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
		StateRoot: Root{},
		BodyRoot:  header.Body.HashTreeRoot(tree.GetHashFn()),
	}
	return state.SetLatestBlockHeader(headerRaw.View())
}
