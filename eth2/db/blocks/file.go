package blocks

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

// FileDB is a simple file-based block database. Each entry is named after 0x prefix block root, with .ssz extension.
// All entries start with a 4 byte fork digest to decode the rest of the block with.
type FileDB struct {
	spec     *common.Spec
	dec      *beacon.ForkDecoder
	basePath string
}

var _ DB = (*FileDB)(nil)

// TODO: refactor to use new Go 1.16 FS type
func NewFileDB(spec *common.Spec, dec *beacon.ForkDecoder, basePath string) *FileDB {
	return &FileDB{spec, dec, basePath}
}

func (db *FileDB) rootToPath(root common.Root) string {
	return path.Join(db.basePath, "0x"+hex.EncodeToString(root[:])+".ssz")
}

// does not overwrite if the file already exists
func (db *FileDB) Store(ctx context.Context, benv *common.BeaconBlockEnvelope) (exists bool, err error) {
	outPath := db.rootToPath(benv.BlockRoot)
	f, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0755)
	defer f.Close()
	if err != nil {
		if os.IsExist(err) {
			return true, nil
		}
		return false, err
	}
	// TODO: add db version byte
	if _, err := f.Write(benv.ForkDigest[:]); err != nil {
		return false, err
	}
	if err := benv.SignedBlock.Serialize(db.spec, codec.NewEncodingWriter(f)); err != nil {
		return false, fmt.Errorf("failed to store block %s: %v", benv.BlockRoot, err)
	}
	return false, nil
}

func (db *FileDB) Import(digest common.ForkDigest, r io.Reader) (exists bool, err error) {
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

func (db *FileDB) Get(ctx context.Context, root common.Root) (envelope *common.BeaconBlockEnvelope, err error) {
	outPath := db.rootToPath(root)
	f, err := os.Open(outPath)
	defer f.Close()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	var digest common.ForkDigest
	if _, err := f.Read(digest[:]); err != nil {
		return nil, err
	}
	return db.dec.DecodeBlock(digest, uint64(info.Size()), f)
}

func (db *FileDB) Size(root common.Root) (size uint64, exists bool) {
	outPath := db.rootToPath(root)
	f, err := os.Open(outPath)
	defer f.Close()
	if err != nil {
		return 0, false
	}
	info, err := f.Stat()
	if err != nil {
		return 0, false
	}
	s := uint64(info.Size())
	if s < 4 {
		// block is corrupt, expected fork digest
		return 0, false
	}
	return s - 4, true
}

func (db *FileDB) Stream(root common.Root) (digest common.ForkDigest, r io.ReadCloser, size uint64, exists bool, err error) {
	outPath := db.rootToPath(root)
	f, err := os.Open(outPath)
	if err != nil {
		return common.ForkDigest{}, nil, 0, false, err

	}
	info, err := f.Stat()
	if err != nil {
		return common.ForkDigest{}, nil, 0, false, err
	}
	if _, err := f.Read(digest[:]); err != nil {
		return common.ForkDigest{}, nil, 0, false, err
	}
	return digest, f, uint64(info.Size()), true, nil
}

func (db *FileDB) Remove(root common.Root) (exists bool, err error) {
	outPath := db.rootToPath(root)
	err = os.Remove(outPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (db *FileDB) Stats() DBStats {
	files, err := ioutil.ReadDir(db.basePath)
	if err != nil {
		return DBStats{}
	}
	// count files, and return latest write, and count of valid looking SSZ block files
	lastMod := common.Root{}
	lastModTime := time.Time{}
	count := 0
	for _, f := range files {
		if name := f.Name(); len(name) == 2+64+4 &&
			strings.HasPrefix(name, "0x") &&
			strings.HasSuffix(name, ".ssz") {
			root, err := hex.DecodeString(name[2 : 2+64])
			if err != nil {
				continue
			}
			if lastModTime.Before(f.ModTime()) {
				lastModTime = f.ModTime()
				copy(lastMod[:], root)
			}
			count += 1
		}
	}
	// return a copy (struct is small and has no pointers)
	return DBStats{
		Count:     0,
		LastWrite: common.Root{},
	}
}

func (db *FileDB) List() (out []common.Root) {
	files, err := ioutil.ReadDir(db.basePath)
	if err != nil {
		return nil
	}
	out = make([]common.Root, 0, len(files))
	for _, f := range files {
		if name := f.Name(); len(name) == 2+64+4 &&
			strings.HasPrefix(name, "0x") &&
			strings.HasSuffix(name, ".ssz") {
			root, err := hex.DecodeString(name[2 : 2+64])
			if err != nil {
				continue
			}
			i := len(out)
			out = append(out, common.Root{})
			copy(out[i][:], root)
		}
	}
	return out
}

func (db *FileDB) Path() string {
	return db.basePath
}

func (db *FileDB) Spec() *common.Spec {
	return db.spec
}
