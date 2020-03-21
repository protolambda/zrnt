package finality

import (
	"errors"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
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

type JustificationFeature struct {
	State FinalityProps
	Meta  interface {
		meta.Versioning
		meta.History
		meta.Staking
		meta.AttesterStatuses
	}
}

func (f *JustificationFeature) Justify(checkpoint Checkpoint) error {
	currentEpoch, err := f.Meta.CurrentEpoch()
	if err != nil {
		return err
	}
	if currentEpoch < checkpoint.Epoch {
		return errors.New("cannot justify future epochs")
	}
	epochsAgo := currentEpoch - checkpoint.Epoch
	if epochsAgo >= Epoch(JUSTIFICATION_BITS_LENGTH) {
		return errors.New("cannot justify history past justification bitfield length")
	}

	if err := f.State.CurrentJustifiedCheckpoint.SetCheckPoint(checkpoint); err != nil {
		return err
	}
	if err := f.State.JustificationBits.SetJustified(epochsAgo); err != nil {
		return err
	}
	return nil
}


var JustificationBitsType = BitvectorType(JUSTIFICATION_BITS_LENGTH)

type JustificationBitsProp BitVectorReadProp

// Prepare bitfield for next epoch by shifting previous bits (truncating to bitfield length)
func (p JustificationBitsProp) NextEpoch() error {
	v, err := BitVectorReadProp(p).BitVector()
	if err != nil {
		return err
	}
	return v.ShRight(1)
}

func (p JustificationBitsProp) IsJustified(epochsAgo ...Epoch) (bool, error) {
	v, err := BitVectorReadProp(p).BitVector()
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
	v, err := BitVectorReadProp(p).BitVector()
	if err != nil {
		return err
	}
	return v.Set(uint64(epochsAgo), true)
}
