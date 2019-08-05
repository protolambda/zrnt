package history

import (
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"github.com/protolambda/zssz"
)

var HistoricalBatchSSZ = zssz.GetSSZ((*HistoricalBatch)(nil))

type HistoricalBatch struct {
	BlockRoots [SLOTS_PER_HISTORICAL_ROOT]Root
	StateRoots [SLOTS_PER_HISTORICAL_ROOT]Root
}

// Return the block root at a recent slot. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (batch *HistoricalBatch) GetBlockRootAtSlot(slot Slot) Root {
	return batch.BlockRoots[slot%SLOTS_PER_HISTORICAL_ROOT]
}

// Return the block root at a recent epoch. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (batch *HistoricalBatch) GetBlockRoot(epoch Epoch) Root {
	return batch.GetBlockRootAtSlot(epoch.GetStartSlot())
}

// roots of HistoricalBatch
type HistoricalRoots []Root

func (_ *HistoricalRoots) Limit() uint64 {
	return HISTORICAL_ROOTS_LIMIT
}

type HistoryState struct {
	HistoricalBatch // embedded BlockRoots and StateRoots
	HistoricalRoots HistoricalRoots
}

func (state *HistoryState) SetRecentRoots(slot Slot, blockRoot Root, stateRoot Root) {
	state.BlockRoots[slot%SLOTS_PER_HISTORICAL_ROOT] = blockRoot
	state.StateRoots[slot%SLOTS_PER_HISTORICAL_ROOT] = stateRoot
}

func (state *HistoryState) UpdateHistoricalRoots() {
	historicalBatch := HistoricalBatch{
		BlockRoots: state.BlockRoots,
		StateRoots: state.StateRoots,
	}

	state.HistoricalRoots = append(state.HistoricalRoots, ssz.HashTreeRoot(historicalBatch, HistoricalBatchSSZ))
}
