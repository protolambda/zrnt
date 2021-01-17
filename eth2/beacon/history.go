package beacon

import (
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

// HistoricalBatchRoots stores roots: a batch of state or block roots.
// It represents a Vector[Root, SLOTS_PER_HISTORICAL_ROOT]
type HistoricalBatchRoots []Root

func (a *HistoricalBatchRoots) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return tree.ReadRoots(dr, (*[]Root)(a), uint64(spec.SLOTS_PER_HISTORICAL_ROOT))
}

func (a HistoricalBatchRoots) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, a)
}

func (a HistoricalBatchRoots) ByteLength(spec *Spec) (out uint64) {
	return uint64(spec.SLOTS_PER_HISTORICAL_ROOT) * 32
}

func (a *HistoricalBatchRoots) FixedLength(spec *Spec) uint64 {
	return uint64(spec.SLOTS_PER_HISTORICAL_ROOT) * 32
}

func (li HistoricalBatchRoots) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(li))
	return hFn.ComplexVectorHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length)
}

type HistoricalBatch struct {
	BlockRoots HistoricalBatchRoots `json:"block_roots" yaml:"block_roots"`
	StateRoots HistoricalBatchRoots `json:"state_roots" yaml:"state_roots"`
}

func (a *HistoricalBatch) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(spec.Wrap(&a.BlockRoots), spec.Wrap(&a.StateRoots))
}

func (a *HistoricalBatch) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return w.FixedLenContainer(spec.Wrap(&a.BlockRoots), spec.Wrap(&a.StateRoots))
}

func (a *HistoricalBatch) ByteLength(spec *Spec) uint64 {
	return uint64(spec.SLOTS_PER_HISTORICAL_ROOT) * 32 * 2
}

func (*HistoricalBatch) FixedLength(spec *Spec) uint64 {
	return uint64(spec.SLOTS_PER_HISTORICAL_ROOT) * 32 * 2
}

func (p *HistoricalBatch) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(spec.Wrap(&p.BlockRoots), spec.Wrap(&p.StateRoots))
}

func (c *Phase0Config) BatchRoots() VectorTypeDef {
	return VectorType(RootType, uint64(c.SLOTS_PER_HISTORICAL_ROOT))
}

type BatchRootsView struct{ *ComplexVectorView }

// Return the root at the given slot. Only valid to SLOTS_PER_HISTORICAL_ROOT slots ago.
func (v *BatchRootsView) GetRoot(slot Slot) (Root, error) {
	i := uint64(slot) % v.VectorLength
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
	startSlot, err := spec.EpochStartSlot(epoch)
	if err != nil {
		return Root{}, err
	}
	return blockRoots.GetRoot(startSlot)
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

func (a *HistoricalRoots) Deserialize(spec *Spec, dr *codec.DecodingReader) error {
	return tree.ReadRootsLimited(dr, (*[]Root)(a), spec.HISTORICAL_ROOTS_LIMIT)
}

func (a HistoricalRoots) Serialize(spec *Spec, w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, a)
}

func (a HistoricalRoots) ByteLength(spec *Spec) (out uint64) {
	return uint64(len(a)) * 32
}

func (a *HistoricalRoots) FixedLength(spec *Spec) uint64 {
	return 0 // it's a list, no fixed length
}

func (li HistoricalRoots) HashTreeRoot(spec *Spec, hFn tree.HashFn) Root {
	length := uint64(len(li))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li[i]
		}
		return nil
	}, length, spec.HISTORICAL_ROOTS_LIMIT)
}

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
