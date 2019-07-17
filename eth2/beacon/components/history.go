package components

import (
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

// roots of HistoricalBatch
type HistoricalRoots []Root

func (_ *HistoricalRoots) Limit() uint64 {
	return HISTORICAL_ROOTS_LIMIT
}

type HistoryState struct {
	BlockRoots        [SLOTS_PER_HISTORICAL_ROOT]Root
	StateRoots        [SLOTS_PER_HISTORICAL_ROOT]Root
	HistoricalRoots   HistoricalRoots
}

var HistoricalBatchSSZ = zssz.GetSSZ((*HistoricalBatch)(nil))

type HistoricalBatch struct {
	BlockRoots [SLOTS_PER_HISTORICAL_ROOT]Root
	StateRoots [SLOTS_PER_HISTORICAL_ROOT]Root
}

// Return the block root at the given slot (a recent one)
func (state *HistoryState) GetBlockRootAtSlot(meta VersioningMeta, slot Slot) (Root, error) {
	currentSlot := meta.Slot()
	if !(slot < currentSlot && slot+SLOTS_PER_HISTORICAL_ROOT <= currentSlot) {
		return Root{}, fmt.Errorf("cannot get block root for given slot %d, current slot is %d", slot, currentSlot)
	}
	return state.BlockRoots[slot%SLOTS_PER_HISTORICAL_ROOT], nil
}

// Return the block root at a recent epoch
func (state *HistoryState) GetBlockRoot(meta VersioningMeta, epoch Epoch) (Root, error) {
	return state.GetBlockRootAtSlot(meta, epoch.GetStartSlot())
}

func (state *HistoryState) UpdateHistoricalRoots() {
	historicalBatch := HistoricalBatch{
		BlockRoots: state.BlockRoots,
		StateRoots: state.StateRoots,
	}

	state.HistoricalRoots = append(state.HistoricalRoots, ssz.HashTreeRoot(historicalBatch, HistoricalBatchSSZ))
}
