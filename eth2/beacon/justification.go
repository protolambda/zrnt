package beacon

import (
	"context"
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

func (spec *Spec) ProcessEpochJustification(ctx context.Context, epc *EpochsContext, process *EpochProcess, state *BeaconStateView) error {
	select {
	case <-ctx.Done():
		return TransitionCancelErr
	default: // Don't block.
		break
	}
	previousEpoch := process.PrevEpoch
	currentEpoch := process.CurrEpoch

	// skip if genesis.
	if currentEpoch <= GENESIS_EPOCH+1 {
		return nil
	}

	prJustCh, err := state.PreviousJustifiedCheckpoint()
	if err != nil {
		return err
	}
	oldPreviousJustified, err := prJustCh.Raw()
	if err != nil {
		return err
	}
	cuJustCh, err := state.CurrentJustifiedCheckpoint()
	if err != nil {
		return err
	}
	oldCurrentJustified, err := cuJustCh.Raw()
	if err != nil {
		return err
	}

	bitsView, err := state.JustificationBits()
	if err != nil {
		return err
	}
	bits, err := bitsView.Raw()
	if err != nil {
		return err
	}

	// Rotate (a copy of) current into previous
	if err := prJustCh.Set(&oldCurrentJustified); err != nil {
		return err
	}

	bits.NextEpoch()

	// Get the total current stake
	totalStake := process.TotalActiveStake

	var newJustifiedCheckpoint *Checkpoint
	// > Justification
	if process.PrevEpochUnslashedStake.TargetStake*3 >= totalStake*2 {
		root, err := spec.GetBlockRoot(state, previousEpoch)
		if err != nil {
			return err
		}
		newJustifiedCheckpoint = &Checkpoint{
			Epoch: previousEpoch,
			Root:  root,
		}
		bits[0] |= 1 << 1
	}
	if process.CurrEpochUnslashedTargetStake*3 >= totalStake*2 {
		root, err := spec.GetBlockRoot(state, currentEpoch)
		if err != nil {
			return err
		}
		newJustifiedCheckpoint = &Checkpoint{
			Epoch: currentEpoch,
			Root:  root,
		}
		bits[0] |= 1 << 0
	}
	if newJustifiedCheckpoint != nil {
		if err := cuJustCh.Set(newJustifiedCheckpoint); err != nil {
			return err
		}
	}

	// > Finalization
	var toFinalize *Checkpoint
	// The 2nd/3rd/4th most recent epochs are all justified, the 2nd using the 4th as source
	if justified := bits.IsJustified(1, 2, 3); justified && oldPreviousJustified.Epoch+3 == currentEpoch {
		toFinalize = &oldPreviousJustified
	}
	// The 2nd/3rd most recent epochs are both justified, the 2nd using the 3rd as source
	if justified := bits.IsJustified(1, 2); justified && oldPreviousJustified.Epoch+2 == currentEpoch {
		toFinalize = &oldPreviousJustified
	}
	// The 1st/2nd/3rd most recent epochs are all justified, the 1st using the 3rd as source
	if justified := bits.IsJustified(0, 1, 2); justified && oldCurrentJustified.Epoch+2 == currentEpoch {
		toFinalize = &oldCurrentJustified
	}
	// The 1st/2nd most recent epochs are both justified, the 1st using the 2nd as source
	if justified := bits.IsJustified(0, 1); justified && oldCurrentJustified.Epoch+1 == currentEpoch {
		toFinalize = &oldCurrentJustified
	}
	if toFinalize != nil {
		finCh, err := state.FinalizedCheckpoint()
		if err != nil {
			return err
		}
		if err := finCh.Set(toFinalize); err != nil {
			return err
		}
	}
	if err := bitsView.Set(bits); err != nil {
		return err
	}
	return nil
}

type JustificationBits [1]byte

func (b *JustificationBits) Deserialize(dr *codec.DecodingReader) error {
	v, err := dr.ReadByte()
	if err != nil {
		return err
	}
	b[0] = v
	return nil
}

func (a JustificationBits) Serialize(w *codec.EncodingWriter) error {
	return w.WriteByte(a[0])
}

func (jb JustificationBits) FixedLength() uint64 {
	return 1
}

func (jb JustificationBits) ByteLength() uint64 {
	return 1
}

func (jb JustificationBits) HashTreeRoot(hFn tree.HashFn) Root {
	return Root{0: jb[0]}
}

func (jb *JustificationBits) BitLen() uint64 {
	return JUSTIFICATION_BITS_LENGTH
}

// Prepare bitfield for next epoch by shifting previous bits (truncating to bitfield length)
func (jb *JustificationBits) NextEpoch() {
	// shift and mask
	jb[0] = (jb[0] << 1) & 0x0f
}

func (jb *JustificationBits) IsJustified(epochsAgo ...Epoch) bool {
	for _, t := range epochsAgo {
		if jb[0]&(1<<t) == 0 {
			return false
		}
	}
	return true
}

var JustificationBitsType = BitVectorType(JUSTIFICATION_BITS_LENGTH)

type JustificationBitsView struct {
	*BitVectorView
}

func (v *JustificationBitsView) Raw() (JustificationBits, error) {
	b, err := v.SubtreeView.GetNode(0)
	if err != nil {
		return JustificationBits{}, err
	}
	r, ok := b.(*Root)
	if !ok {
		return JustificationBits{}, fmt.Errorf("justification bitvector bottom node is not a root, cannot get bits")
	}
	return JustificationBits{r[0]}, nil
}

func (v *JustificationBitsView) Set(bits JustificationBits) error {
	root := Root{0: bits[0]}
	return v.SetBacking(&root)
}

func AsJustificationBits(v View, err error) (*JustificationBitsView, error) {
	c, err := AsBitVector(v, err)
	return &JustificationBitsView{c}, err
}
