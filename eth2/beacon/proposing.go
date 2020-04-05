package beacon

import (
	"encoding/binary"
	"errors"
	"fmt"


	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/shuffle"
)

type EpochProposerIndices [SLOTS_PER_EPOCH]ValidatorIndex

type ProposersData struct {
	Epoch   Epoch
	Current *EpochProposerIndices
	Next    *EpochProposerIndices
}

func (state *ProposersData) GetBeaconProposerIndex(slot Slot) (ValidatorIndex, error) {
	epoch := slot.ToEpoch()
	if epoch == state.Epoch {
		return state.Current[slot-state.Epoch.GetStartSlot()], nil
	} else if epoch == state.Epoch+1 {
		return state.Next[slot-(state.Epoch + 1).GetStartSlot()], nil
	} else {
		return 0, fmt.Errorf("slot %d not within range", slot)
	}
}

func LoadBeaconProposersData(input PrepareProposersInput) (out *ProposersData, err error) {
	currentEpoch, err := input.CurrentEpoch()
	if err != nil {
		return nil, err
	}
	curr, err := LoadBeaconProposerIndices(input, currentEpoch)
	if err != nil {
		return nil, err
	}
	next, err := LoadBeaconProposerIndices(input, currentEpoch + 1)
	return &ProposersData{
		Epoch:   currentEpoch,
		Current: curr,
		Next:    next,
	}, nil
}


// Return the beacon proposer index for the current slot
func LoadBeaconProposerIndices(input PrepareProposersInput, epoch Epoch) (out *EpochProposerIndices, err error) {
	seedSource, err := input.GetSeed(epoch, DOMAIN_BEACON_PROPOSER)
	if err != nil {
		return nil, err
	}
	indices, err := input.GetActiveValidatorIndices(epoch)
	if err != nil {
		return nil, err
	}

	out = new(EpochProposerIndices)

	startSlot := epoch.GetStartSlot()
	for i := Slot(0); i < SLOTS_PER_EPOCH; i++ {
		buf := make([]byte, 32+8, 32+8)
		copy(buf[0:32], seedSource[:])
		binary.LittleEndian.PutUint64(buf[32:], uint64(startSlot+i))
		seed := Hash(buf)
		proposerIndex, err := computeProposerIndex(input, indices, seed)
		if err != nil {
			return nil, err
		}
		out[i] = proposerIndex
	}
	return
}
