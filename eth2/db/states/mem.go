package states

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/tree"
	"sync"
)

type MemDB struct {
	// beacon.Root -> tree.Node (backing of BeaconStateView)
	data sync.Map
	spec *common.Spec
}

func NewMemDB(spec *common.Spec) *MemDB {
	return &MemDB{spec: spec}
}

func (db *MemDB) Store(ctx context.Context, state common.BeaconState) error {
	// Released when the block is removed from the DB
	root := state.HashTreeRoot(tree.GetHashFn())
	db.data.Store(root, state)
	return nil
}

func (db *MemDB) Get(ctx context.Context, root common.Root) (state common.BeaconState, err error) {
	dat, ok := db.data.Load(root)
	if !ok {
		return nil, nil
	}
	state, ok = dat.(common.BeaconState)
	if !ok {
		panic("in-memory db was corrupted with unexpected state type")
	}
	return
}

func (db *MemDB) Remove(root common.Root) error {
	db.data.Delete(root)
	return nil
}

func (db *MemDB) Close() error {
	return nil
}
