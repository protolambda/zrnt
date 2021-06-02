package states

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/tree"
	"sync"
)

type MemDB struct {
	// beacon.Root -> tree.Node (backing of BeaconStateView)
	data        sync.Map
	removalLock sync.Mutex
	spec        *common.Spec
}

func NewMemDB(spec *common.Spec) *MemDB {
	return &MemDB{spec: spec}
}

func (db *MemDB) Store(ctx context.Context, state common.BeaconState) (exists bool, err error) {
	// Released when the block is removed from the DB
	root := state.HashTreeRoot(tree.GetHashFn())
	_, loaded := db.data.LoadOrStore(root, state)
	return loaded, nil
}

func (db *MemDB) Get(ctx context.Context, root common.Root) (state common.BeaconState, exists bool, err error) {
	dat, ok := db.data.Load(root)
	if !ok {
		return nil, false, nil
	}
	exists = true
	state, ok = dat.(common.BeaconState)
	if !ok {
		panic("in-memory db was corrupted with unexpected state type")
	}
	return
}

func (db *MemDB) Remove(root common.Root) (exists bool, err error) {
	db.removalLock.Lock()
	defer db.removalLock.Unlock()
	_, ok := db.data.Load(root)
	db.data.Delete(root)
	return ok, nil
}
