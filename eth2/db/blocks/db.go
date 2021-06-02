package blocks

import (
	"bytes"
	"context"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"io"
	"sync"
)

type DBStats struct {
	Count     int64
	LastWrite common.Root
}

type DB interface {
	// Store, only for trusted blocks, to persist a block in the DB.
	// The block is stored in serialized form, so the original instance may be mutated after storing it.
	// This is an efficient convenience method for using Import.
	// Returns exists=true if the block exists (previously), false otherwise. If error, it may not be accurate.
	// Returns slashable=true if exists=true, but the signatures are different. The existing block is kept.
	Store(ctx context.Context, benv *common.BeaconBlockEnvelope) (exists bool, err error)
	// Import inserts a SignedBeaconBlock, read directly from the reader stream.
	// Returns exists=true if the block exists (previously), false otherwise. If error, it may not be accurate.
	// Returns slashable=true if exists=true, but the signatures are different. The existing block is kept.
	Import(digest common.ForkDigest, r io.Reader) (exists bool, err error)
	// Get, a convenience method for getting a block. The block is safe to modify.
	// Returns the envelope if the block exists, nil otherwise. If error, exists-check may not be accurate.
	Get(ctx context.Context, root common.Root) (envelope *common.BeaconBlockEnvelope, err error)
	// Size quickly checks the size of a block, without dealing with the full block.
	// Returns exists=true if the block exists, false otherwise. If error, it may not be accurate.
	Size(root common.Root) (size uint64, exists bool)
	// Stream is used to stream the contents by getting a reader and total size to read
	Stream(root common.Root) (digest common.ForkDigest, r io.ReadCloser, size uint64, exists bool, err error)
	// Remove removes a block from the DB. Removing a block that does not exist is safe.
	// Returns exists=true if the block exists (previously), false otherwise. If error, it may not be accurate.
	Remove(root common.Root) (exists bool, err error)
	// Stats shows some database statistics such as latest write key and entry count.
	Stats() DBStats
	// List all known block roots
	List() []common.Root
	// Get Path
	Path() string
	// Spec of blocks
	Spec() *common.Spec
	io.Closer
}

// Mainnet blocks are 157756 in v0.12.x, buffer can grow if necessary, and should be enough for most custom configs.
var maxBlockSize = 160_000

var dbBlockPool = sync.Pool{
	New: func() interface{} {
		// ensure enough capacity for any block. We pool it anyway, so eventually it may grow that big.
		return bytes.NewBuffer(make([]byte, 0, maxBlockSize))
	},
}

func getPoolBlockBuf() *bytes.Buffer {
	buf := dbBlockPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}
