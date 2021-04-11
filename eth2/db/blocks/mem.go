package blocks

import (
	"bytes"
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
	"github.com/protolambda/ztyp/codec"
	"io"
	"sync"
	"sync/atomic"
)

type MemDB struct {
	// beacon.Root -> []byte (fork digest ++ serialized SignedBeaconBlock)
	data        sync.Map
	removalLock sync.Mutex
	stats       DBStats
	dec         *beacon.ForkDecoder
	spec        *common.Spec
}

var _ DB = (*MemDB)(nil)

func NewMemDB(spec *common.Spec, dec *beacon.ForkDecoder) *MemDB {
	return &MemDB{spec: spec, dec: dec}
}

func (db *MemDB) Store(ctx context.Context, benv *common.BeaconBlockEnvelope) (exists bool, err error) {
	// Released when the block is removed from the DB
	buf := getPoolBlockBuf()
	if _, err := buf.Write(benv.ForkDigest[:]); err != nil {
		return false, err
	}
	err = benv.SignedBlock.Serialize(db.spec, codec.NewEncodingWriter(buf))
	if err != nil {
		return false, fmt.Errorf("failed to store block %s: %v", benv.BlockRoot, err)
	}
	existing, loaded := db.data.LoadOrStore(benv.BlockRoot, buf.Bytes())
	if loaded {
		existingBlock := existing.(*phase0.SignedBeaconBlock)
		sigDifference := existingBlock.Signature != benv.Signature
		dbBlockPool.Put(buf) // put it back, we didn't store it
		if sigDifference {
			return true, fmt.Errorf("block %s already exists, but its signature %x does not match new signature %s",
				benv.BlockRoot, existingBlock.Signature, benv.Signature)
		}
	} else {
		atomic.AddInt64(&db.stats.Count, 1)
		db.stats.LastWrite = benv.BlockRoot
	}
	return loaded, nil
}

func (db *MemDB) Import(digest common.ForkDigest, r io.Reader) (exists bool, err error) {
	buf := getPoolBlockBuf()
	defer dbBlockPool.Put(buf)
	if _, err := buf.ReadFrom(r); err != nil {
		return false, err
	}
	benv, err := db.dec.DecodeBlock(digest, uint64(len(buf.Bytes())), buf)
	if err != nil {
		return false, fmt.Errorf("failed to decode block, nee valid block to get block root. Err: %v", err)
	}
	return db.Store(context.Background(), benv)
}

func (db *MemDB) Get(ctx context.Context, root common.Root) (envelope *common.BeaconBlockEnvelope, err error) {
	dat, ok := db.data.Load(root)
	if !ok {
		return nil, nil
	}
	buf := dat.(*bytes.Buffer)
	var digest common.ForkDigest
	if _, err := buf.Read(digest[:]); err != nil {
		return nil, err
	}
	return db.dec.DecodeBlock(digest, uint64(len(buf.Bytes())), buf)
}

func (db *MemDB) Size(root common.Root) (size uint64, exists bool) {
	dat, ok := db.data.Load(root)
	if !ok {
		return 0, false
	}
	buf := dat.(*bytes.Buffer)
	s := uint64(len(buf.Bytes()))
	if s < 4 {
		// block is corrupt, expected fork digest
		return 0, false
	}
	return s - 4, true
}

type noClose struct {
	io.Reader
}

func (n noClose) Close() error {
	return nil
}

func (db *MemDB) Stream(root common.Root) (digest common.ForkDigest, r io.ReadCloser, size uint64, exists bool, err error) {
	dat, ok := db.data.Load(root)
	if !ok {
		return common.ForkDigest{}, nil, 0, false, nil
	}
	buf := dat.(*bytes.Buffer)
	if _, err := buf.Read(digest[:]); err != nil {
		return common.ForkDigest{}, nil, 0, false, err
	}
	return digest, noClose{buf}, uint64(buf.Len()), true, nil
}

func (db *MemDB) Remove(root common.Root) (exists bool, err error) {
	db.removalLock.Lock()
	defer db.removalLock.Unlock()
	v, ok := db.data.Load(root)
	if ok {
		dbBlockPool.Put(v) // release it back to pool, it's not used in the DB anymore.
		atomic.AddInt64(&db.stats.Count, -1)
	}
	db.data.Delete(root)
	return ok, nil
}

func (db *MemDB) Stats() DBStats {
	// return a copy (struct is small and has no pointers)
	return db.stats
}

func (db *MemDB) List() (out []common.Root) {
	out = make([]common.Root, 0, db.stats.Count)
	db.data.Range(func(key, value interface{}) bool {
		id := key.(common.Root)
		out = append(out, id)
		return true
	})
	return out
}

func (db *MemDB) Path() string {
	return ""
}

func (db *MemDB) Spec() *common.Spec {
	return db.spec
}
