package beacon

import (
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

// Vector[Root, SLOTS_PER_HISTORICAL_ROOT]
type HistoricalBatchRoots []Root

type HistoricalBatch struct {
	BlockRoots HistoricalBatchRoots
	StateRoots HistoricalBatchRoots
}

func (c *Phase0Config) BatchRoots() VectorTypeDef {
	return VectorType(RootType, uint64(c.SLOTS_PER_HISTORICAL_ROOT))
}

type BatchRootsView struct{ *ComplexVectorView }

// Return the root at the given slot. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (v *BatchRootsView) GetRoot(slot Slot) (Root, error) {
	i := uint64(slot) & v.VectorLength
	return AsRoot(v.Get(i))
}

func (v *BatchRootsView) SetRoot(slot Slot, r Root) error {
	i := uint64(slot) % v.VectorLength
	rv := RootView(r)
	return v.Set(i, &rv)
}

func AsBatchRoots(v View, err error) (*BatchRootsView, error) {
	c, err := AsComplexVector(v, err)
	return &BatchRootsView{c}, err
}

// Return the block root at a recent slot. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (spec *Spec) GetBlockRootAtSlot(state *BeaconStateView, slot Slot) (Root, error) {
	blockRoots, err := state.BlockRoots()
	if err != nil {
		return Root{}, err
	}
	return blockRoots.GetRoot(slot)
}

// Return the block root at a recent epoch. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (spec *Spec) GetBlockRoot(state *BeaconStateView, epoch Epoch) (Root, error) {
	blockRoots, err := state.BlockRoots()
	if err != nil {
		return Root{}, err
	}
	return blockRoots.GetRoot(spec.EpochStartSlot(epoch))
}

func (c *Phase0Config) HistoricalBatch() *ContainerTypeDef {
	return ContainerType("HistoricalBatch", []FieldDef{
		{"block_roots", c.BatchRoots()},
		{"state_roots", c.BatchRoots()},
	})
}

type HistoricalBatchView struct{ *ContainerView }

func (v *HistoricalBatchView) BlockRoots() (*BatchRootsView, error) {
	return AsBatchRoots(v.Get(0))
}

func (v *HistoricalBatchView) StateRoots() (*BatchRootsView, error) {
	return AsBatchRoots(v.Get(1))
}

func AsHistoricalBatch(v View, err error) (*HistoricalBatchView, error) {
	c, err := AsContainer(v, err)
	return &HistoricalBatchView{c}, err
}

// roots of HistoricalBatch
type HistoricalRoots []Root

func (c *Phase0Config) HistoricalRoots() ListTypeDef {
	return ListType(RootType, c.HISTORICAL_ROOTS_LIMIT)
}

// roots of HistoricalBatch
type HistoricalRootsView struct{ *ComplexListView }

func AsHistoricalRoots(v View, err error) (*HistoricalRootsView, error) {
	c, err := AsComplexList(v, err)
	return &HistoricalRootsView{c}, err
}

func (state *BeaconStateView) SetRecentRoots(slot Slot, blockRoot Root, stateRoot Root) error {
	blockRootsBatch, err := state.BlockRoots()
	if err != nil {
		return err
	}
	stateRootsBatch, err := state.StateRoots()
	if err != nil {
		return err
	}
	if err := blockRootsBatch.SetRoot(slot%Slot(blockRootsBatch.VectorLength), blockRoot); err != nil {
		return err
	}
	if err := stateRootsBatch.SetRoot(slot%Slot(stateRootsBatch.VectorLength), stateRoot); err != nil {
		return err
	}
	return nil
}

func (state *BeaconStateView) UpdateHistoricalRoots() error {
	histRoots, err := state.HistoricalRoots()
	if err != nil {
		return err
	}
	blockRoots, err := state.BlockRoots()
	if err != nil {
		return err
	}
	stateRoots, err := state.StateRoots()
	if err != nil {
		return err
	}
	// emulating HistoricalBatch here
	hFn := tree.GetHashFn()
	newHistoricalRoot := RootView(tree.Hash(blockRoots.HashTreeRoot(hFn), stateRoots.HashTreeRoot(hFn)))
	return histRoots.Append(&newHistoricalRoot)
}
