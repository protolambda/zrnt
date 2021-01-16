package states

import (
	"context"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/ztyp/tree"
	"sync"
	"sync/atomic"
)

type MemDB struct {
	// beacon.Root -> tree.Node (backing of BeaconStateView)
	data        sync.Map
	removalLock sync.Mutex
	stats       DBStats
	spec        *beacon.Spec
}

func NewMemDB(spec *beacon.Spec) *MemDB {
	return &MemDB{spec: spec}
}

func (db *MemDB) Store(ctx context.Context, state *beacon.BeaconStateView) (exists bool, err error) {
	// Released when the block is removed from the DB
	root := state.HashTreeRoot(tree.GetHashFn())
	backing := state.Backing()
	_, loaded := db.data.LoadOrStore(root, backing)
	if !loaded {
		atomic.AddInt64(&db.stats.Count, 1)
		db.stats.LastWrite = root
	}
	return loaded, nil
}

func (db *MemDB) Get(ctx context.Context, root beacon.Root) (state *beacon.BeaconStateView, exists bool, err error) {
	dat, ok := db.data.Load(root)
	if !ok {
		return nil, false, nil
	}
	exists = true
	v, vErr := db.spec.BeaconState().ViewFromBacking(dat.(tree.Node), nil)
	state, err = beacon.AsBeaconStateView(v, vErr)
	return
}

func (db *MemDB) Remove(root beacon.Root) (exists bool, err error) {
	db.removalLock.Lock()
	defer db.removalLock.Unlock()
	_, ok := db.data.Load(root)
	if ok {
		atomic.AddInt64(&db.stats.Count, -1)
	}
	db.data.Delete(root)
	return ok, nil
}

func (db *MemDB) Stats() DBStats {
	// return a copy (struct is small and has no pointers)
	return db.stats
}

func (db *MemDB) List() (out []beacon.Root) {
	out = make([]beacon.Root, 0, db.stats.Count)
	db.data.Range(func(key, value interface{}) bool {
		id := key.(beacon.Root)
		out = append(out, id)
		return true
	})
	return out
}

func (db *MemDB) Path() string {
	return ""
}

func (db *MemDB) Spec() *beacon.Spec {
	return db.spec
}
