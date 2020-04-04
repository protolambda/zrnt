package history

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var BatchRootsType = VectorType(RootType, uint64(SLOTS_PER_HISTORICAL_ROOT))

type BatchRoots struct{ *ComplexVectorView }

func (br *BatchRoots) GetRoot(slot Slot) (Root, error) {
	return RootReadProp(PropReader(br, uint64(slot%SLOTS_PER_HISTORICAL_ROOT))).Root()
}

func (br *BatchRoots) SetRoot(slot Slot, v Root) error {
	return RootWriteProp(PropWriter(br, uint64(slot%SLOTS_PER_HISTORICAL_ROOT))).SetRoot(v)
}

type BatchRootsProp ComplexVectorProp

func (p BatchRootsProp) BatchRoots() (*BatchRoots, error) {
	v, err := ComplexVectorProp(p).Vector()
	if err != nil {
		return nil, err
	}
	return &BatchRoots{ComplexVectorView: v}, nil
}

type BlockRootsProp BatchRootsProp

func (p BlockRootsProp) BlockRoots() (*BatchRoots, error) {
	return BatchRootsProp(p).BatchRoots()
}

// Return the block root at a recent slot. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (p BlockRootsProp) GetBlockRootAtSlot(slot Slot) (Root, error) {
	batch, err := p.BlockRoots()
	if err != nil {
		return Root{}, err
	}
	return batch.GetRoot(slot % SLOTS_PER_HISTORICAL_ROOT)
}

// Return the block root at a recent epoch. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (p BlockRootsProp) GetBlockRoot(epoch Epoch) (Root, error) {
	return p.GetBlockRootAtSlot(epoch.GetStartSlot())
}

type StateRootsProp BatchRootsProp

func (p StateRootsProp) StateRoots() (*BatchRoots, error) {
	return BatchRootsProp(p).BatchRoots()
}

type HistoricalBatch struct{ *ContainerView }

func (hb *HistoricalBatch) BlockRoots() (*BatchRoots, error) {
	return BlockRootsProp(PropReader(hb, 0)).BlockRoots()
}

func (hb *HistoricalBatch) StateRoots() (*BatchRoots, error) {
	return StateRootsProp(PropReader(hb, 1)).StateRoots()
}

var HistoricalBatchType = ContainerType("HistoricalBatch", []FieldDef{
	{"block_roots", BatchRootsType},
	{"state_roots", BatchRootsType},
})

// roots of HistoricalBatch
type HistoricalRoots struct{ *ComplexListView }

var HistoricalRootsType = ListType(RootType, HISTORICAL_ROOTS_LIMIT)

type HistoricalRootsProp ComplexListProp

func (p HistoricalRootsProp) HistoricalRoots() (*HistoricalRoots, error) {
	v, err := ComplexListProp(p).List()
	if v != nil {
		return nil, err
	}
	return &HistoricalRoots{ComplexListView: v}, nil
}

type HistoryProps struct {
	BlockRootsProp
	StateRootsProp
	HistoricalRootsProp
}

func (p HistoryProps) SetRecentRoots(slot Slot, blockRoot Root, stateRoot Root) error {
	blockRootsBatch, err := p.BlockRoots()
	if err != nil {
		return err
	}
	stateRootsBatch, err := p.StateRoots()
	if err != nil {
		return err
	}
	if err := blockRootsBatch.SetRoot(slot%SLOTS_PER_HISTORICAL_ROOT, blockRoot); err != nil {
		return err
	}
	if err := stateRootsBatch.SetRoot(slot%SLOTS_PER_HISTORICAL_ROOT, stateRoot); err != nil {
		return err
	}
	return nil
}

func (p *HistoryProps) UpdateHistoricalRoots() error {
	histRoots, err := p.HistoricalRoots()
	if err != nil {
		return err
	}
	blockRoots, err := p.BlockRoots()
	if err != nil {
		return err
	}
	stateRoots, err := p.StateRoots()
	if err != nil {
		return err
	}
	// emulating HistoricalBatch here
	hFn := tree.GetHashFn()
	newHistoricalRoot := RootView(tree.Hash(blockRoots.HashTreeRoot(hFn), stateRoots.HashTreeRoot(hFn)))
	return histRoots.Append(&newHistoricalRoot)
}
