package beacon

import (
	"errors"

	. "github.com/protolambda/ztyp/props"
	. "github.com/protolambda/ztyp/view"
)

type FinalityProps struct {
	JustificationBits           JustificationBitsProp
	PreviousJustifiedCheckpoint CheckpointProp
	CurrentJustifiedCheckpoint  CheckpointProp
	FinalizedCheckpoint         CheckpointProp
}

func (state *FinalityProps) Finalized() (Checkpoint, error) {
	return state.FinalizedCheckpoint.CheckPoint()
}

func (state *FinalityProps) PreviousJustified() (Checkpoint, error) {
	return state.PreviousJustifiedCheckpoint.CheckPoint()
}

func (state *FinalityProps) CurrentJustified() (Checkpoint, error) {
	return state.CurrentJustifiedCheckpoint.CheckPoint()
}

func (state *FinalityProps) Justify(currentEpoch Epoch, checkpoint Checkpoint) error {
	if currentEpoch < checkpoint.Epoch {
		return errors.New("cannot justify future epochs")
	}
	epochsAgo := currentEpoch - checkpoint.Epoch
	if epochsAgo >= Epoch(JUSTIFICATION_BITS_LENGTH) {
		return errors.New("cannot justify history past justification bitfield length")
	}

	if err := state.CurrentJustifiedCheckpoint.SetCheckPoint(checkpoint); err != nil {
		return err
	}
	if err := state.JustificationBits.SetJustified(epochsAgo); err != nil {
		return err
	}
	return nil
}


var JustificationBitsType = BitVectorType(JUSTIFICATION_BITS_LENGTH)

type JustificationBitsProp BitVectorProp

// Prepare bitfield for next epoch by shifting previous bits (truncating to bitfield length)
func (p JustificationBitsProp) NextEpoch() error {
	v, err := BitVectorProp(p).BitVector()
	if err != nil {
		return err
	}
	return v.ShRight(1)
}

func (p JustificationBitsProp) IsJustified(epochsAgo ...Epoch) (bool, error) {
	v, err := BitVectorProp(p).BitVector()
	if err != nil {
		return false, err
	}
	for _, t := range epochsAgo {
		if bit, err := v.Get(uint64(t)); err != nil {
			return false, err
		} else if !bit {
			return false, nil
		}
	}
	return true, nil
}

func (p JustificationBitsProp) SetJustified(epochsAgo Epoch) error {
	v, err := BitVectorProp(p).BitVector()
	if err != nil {
		return err
	}
	return v.Set(uint64(epochsAgo), true)
}
