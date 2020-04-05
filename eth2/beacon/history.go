package beacon

import (
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

var BlockRootsType = VectorType(RootType, uint64(SLOTS_PER_HISTORICAL_ROOT))

type BlockRootsView struct{ *ComplexVectorView }

// Return the block root at the given slot. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (v *BlockRootsView) GetRoot(slot Slot) (Root, error) {
	i := uint64(slot%SLOTS_PER_HISTORICAL_ROOT)
	return AsRoot(v.Get(i))
}

func (v *BlockRootsView) SetRoot(slot Slot, r Root) error {
	i := uint64(slot%SLOTS_PER_HISTORICAL_ROOT)
	rv := RootView(r)
	return v.Set(i, &rv)
}

func AsBlockRoots(v View, err error) (*BlockRootsView, error) {
	c, err := AsComplexVector(v, err)
	return &BlockRootsView{c}, err
}

// Return the block root at a recent slot. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (state *BeaconStateView) GetBlockRootAtSlot(slot Slot) (Root, error) {
	blockRoots, err := state.BlockRoots()
	if err != nil {
		return Root{}, err
	}
	return blockRoots.GetRoot(slot)
}

// Return the block root at a recent epoch. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (state *BeaconStateView) GetBlockRoot(epoch Epoch) (Root, error) {
	blockRoots, err := state.BlockRoots()
	if err != nil {
		return Root{}, err
	}
	return blockRoots.GetRoot(epoch.GetStartSlot())
}

var StateRootsType = VectorType(RootType, uint64(SLOTS_PER_HISTORICAL_ROOT))

type StateRootsView struct{ *ComplexVectorView }

// Return the state root at the given slot. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (v *StateRootsView) GetRoot(slot Slot) (Root, error) {
	i := uint64(slot%SLOTS_PER_HISTORICAL_ROOT)
	return AsRoot(v.Get(i))
}

func (v *StateRootsView) SetRoot(slot Slot, r Root) error {
	i := uint64(slot%SLOTS_PER_HISTORICAL_ROOT)
	rv := RootView(r)
	return v.Set(i, &rv)
}

func AsStateRoots(v View, err error) (*StateRootsView, error) {
	c, err := AsComplexVector(v, err)
	return &StateRootsView{c}, err
}

var HistoricalBatchType = ContainerType("HistoricalBatch", []FieldDef{
	{"block_roots", BlockRootsType},
	{"state_roots", StateRootsType},
})

type HistoricalBatchView struct{ *ContainerView }

func (v *HistoricalBatchView) BlockRoots() (*BlockRootsView, error) {
	return AsBlockRoots(v.Get(0))
}

func (v *HistoricalBatchView) StateRoots() (*StateRootsView, error) {
	return AsStateRoots(v.Get(1))
}

func AsHistoricalBatch(v View, err error) (*HistoricalBatchView, error) {
	c, err := AsContainer(v, err)
	return &HistoricalBatchView{c}, err
}

var HistoricalRootsType = ListType(RootType, HISTORICAL_ROOTS_LIMIT)

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
	if err := blockRootsBatch.SetRoot(slot%SLOTS_PER_HISTORICAL_ROOT, blockRoot); err != nil {
		return err
	}
	if err := stateRootsBatch.SetRoot(slot%SLOTS_PER_HISTORICAL_ROOT, stateRoot); err != nil {
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
