package history

import (
	. "github.com/protolambda/zrnt/eth2/core"
	. "github.com/protolambda/ztyp/props"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var BatchRootsType = VectorType(RootType, uint64(SLOTS_PER_HISTORICAL_ROOT))

type BatchRoots struct{ *VectorView }

func (br *BatchRoots) GetRoot(slot Slot) (Root, error) {
	return RootReadProp(PropReader(br, uint64(slot%SLOTS_PER_HISTORICAL_ROOT))).Root()
}

func (br *BatchRoots) SetRoot(slot Slot, v Root) error {
	return RootWriteProp(PropWriter(br, uint64(slot%SLOTS_PER_HISTORICAL_ROOT))).SetRoot(v)
}

type BatchRootsReadProp VectorReadProp

func (p BatchRootsReadProp) BatchRoots() (*BatchRoots, error) {
	v, err := VectorReadProp(p).Vector()
	if err != nil {
		return nil, err
	}
	return &BatchRoots{VectorView: v}, nil
}

type BatchRootsWriteProp VectorWriteProp

func (p BatchRootsWriteProp) SetBatchRoots(v *BatchRoots) error {
	return VectorWriteProp(p).SetVector(v.VectorView)
}

type BlockRootsReadProp BatchRootsReadProp

func (p BlockRootsReadProp) BlockRoots() (*BatchRoots, error) {
	return BatchRootsReadProp(p).BatchRoots()
}

// Return the block root at a recent slot. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (p BlockRootsReadProp) GetBlockRootAtSlot(slot Slot) (Root, error) {
	batch, err := p.BlockRoots()
	if err != nil {
		return Root{}, err
	}
	return batch.GetRoot(slot % SLOTS_PER_HISTORICAL_ROOT)
}

// Return the block root at a recent epoch. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (p BlockRootsReadProp) GetBlockRoot(epoch Epoch) (Root, error) {
	return p.GetBlockRootAtSlot(epoch.GetStartSlot())
}

type BlockRootsWriteProp BatchRootsWriteProp

func (p BlockRootsWriteProp) SetBlockRoots(v *BatchRoots) error {
	return BatchRootsWriteProp(p).SetBatchRoots(v)
}

type StateRootsReadProp BatchRootsReadProp

func (p StateRootsReadProp) StateRoots() (*BatchRoots, error) {
	return BatchRootsReadProp(p).BatchRoots()
}

type StateRootsWriteProp BatchRootsWriteProp

func (p StateRootsWriteProp) SetStateRoots(v *BatchRoots) error {
	return BatchRootsWriteProp(p).SetBatchRoots(v)
}

type HistoricalBatch struct{ *ContainerView }

func (hb *HistoricalBatch) BlockRoots() (*BatchRoots, error) {
	return BlockRootsReadProp(PropReader(hb, 0)).BlockRoots()
}

func (hb *HistoricalBatch) SetBlockRoots(v *BatchRoots) error {
	return BlockRootsWriteProp(PropWriter(hb, 0)).SetBlockRoots(v)
}

func (hb *HistoricalBatch) StateRoots() (*BatchRoots, error) {
	return StateRootsReadProp(PropReader(hb, 1)).StateRoots()
}

func (hb *HistoricalBatch) SetStateRoots(v *BatchRoots) error {
	return StateRootsWriteProp(PropWriter(hb, 1)).SetStateRoots(v)
}

var HistoricalBatchType = &ContainerType{
	{"block_roots", BatchRootsType},
	{"state_roots", BatchRootsType},
}

type HistoricalBatchReadProp ContainerReadProp

func (p HistoricalBatchReadProp) HistoricalBatch() (*HistoricalBatch, error) {
	v, err := ContainerReadProp(p).Container()
	if err != nil {
		return nil, err
	}
	return &HistoricalBatch{ContainerView: v}, nil
}

// roots of HistoricalBatch
type HistoricalRoots struct{ *ListView }

var HistoricalRootsType = ListType(RootType, HISTORICAL_ROOTS_LIMIT)

type HistoricalRootsReadProp ListReadProp

func (p HistoricalRootsReadProp) HistoricalRoots() (*HistoricalRoots, error) {
	v, err := ListReadProp(p).List()
	if v != nil {
		return nil, err
	}
	return &HistoricalRoots{ListView: v}, nil
}

type HistoricalRootsWriteProp ListWriteProp

func (p HistoricalRootsWriteProp) SetHistoricalRoots(v *HistoricalRoots) error {
	return p(v)
}

type HistoryMutProps struct {
	BlockRootsReadProp
	StateRootsReadProp
	BlockRootsWriteProp
	StateRootsWriteProp
	HistoricalRootsReadProp
	HistoricalRootsWriteProp
}

func (p HistoryMutProps) SetRecentRoots(slot Slot, blockRoot Root, stateRoot Root) error {
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

func (p *HistoryMutProps) UpdateHistoricalRoots() error {
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
	batch := &HistoricalBatch{ContainerView: HistoricalBatchType.New()}
	if err := batch.SetBlockRoots(blockRoots); err != nil {
		return err
	}
	if err := batch.SetStateRoots(stateRoots); err != nil {
		return err
	}
	newHistoricalRoot := batch.ViewRoot(tree.Hash)
	if err := histRoots.Append(&newHistoricalRoot); err != nil {
		return err
	}
	return p.SetHistoricalRoots(histRoots)
}
