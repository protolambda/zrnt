package blocks

import (
	"bytes"
	"context"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/ztyp/tree"
	"io"
	"sync"
)

type BlockWithRoot struct {
	// Root of the Block.Message
	Root beacon.Root
	// Block, with signature
	Block *beacon.SignedBeaconBlock
}

func WithRoot(spec *beacon.Spec, block *beacon.SignedBeaconBlock) *BlockWithRoot {
	root := block.Message.HashTreeRoot(spec, tree.GetHashFn())
	return &BlockWithRoot{
		Root:  root,
		Block: block,
	}
}

type DBStats struct {
	Count     int64
	LastWrite beacon.Root
}

type DB interface {
	// Store, only for trusted blocks, to persist a block in the DB.
	// The block is stored in serialized form, so the original instance may be mutated after storing it.
	// This is an efficient convenience method for using Import.
	// Returns exists=true if the block exists (previously), false otherwise. If error, it may not be accurate.
	// Returns slashable=true if exists=true, but the signatures are different. The existing block is kept.
	Store(ctx context.Context, block *BlockWithRoot) (exists bool, err error)
	// Import inserts a SignedBeaconBlock, read directly from the reader stream.
	// Returns exists=true if the block exists (previously), false otherwise. If error, it may not be accurate.
	// Returns slashable=true if exists=true, but the signatures are different. The existing block is kept.
	Import(r io.Reader) (exists bool, err error)
	// Get, an efficient convenience method for getting a block through Export. The block is safe to modify.
	// The data at the pointer is mutated to the new block.
	// Returns exists=true if the block exists, false otherwise. If error, it may not be accurate.
	Get(ctx context.Context, root beacon.Root, dest *beacon.SignedBeaconBlock) (exists bool, err error)
	// Size quickly checks the size of a block, without dealing with the full block.
	// Returns exists=true if the block exists, false otherwise. If error, it may not be accurate.
	Size(root beacon.Root) (size uint64, exists bool)
	// Export outputs the requested SignedBeaconBlock to the writer in SSZ.
	// Returns exists=true if the block exists, false otherwise. If error, it may not be accurate.
	Export(root beacon.Root, w io.Writer) (exists bool, err error)
	// Stream is used to stream the contents by getting a reader and total size to read
	Stream(root beacon.Root) (r io.ReadCloser, size uint64, exists bool, err error)
	// Remove removes a block from the DB. Removing a block that does not exist is safe.
	// Returns exists=true if the block exists (previously), false otherwise. If error, it may not be accurate.
	Remove(root beacon.Root) (exists bool, err error)
	// Stats shows some database statistics such as latest write key and entry count.
	Stats() DBStats
	// List all known block roots
	List() []beacon.Root
	// Get Path
	Path() string
	// Spec of states
	Spec() *beacon.Spec
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
