package common

import (
	"fmt"
	"sync"
)

// BuilderPubkeyCache is shared between any state. However, if .Append(index, pubkey) conflicts, a new cache will be forked out.
type BuilderPubkeyCache struct {
	parent *BuilderPubkeyCache
	// The count up until the conflicting builder index (cause of the pubkey cache fork).
	trustedParentCount BuilderIndex
	pub2idx            map[BLSPubkey]BuilderIndex
	// starting at trustedParentCount
	idx2pub []CachedPubkey
	// Can have many reads concurrently, but only 1 write.
	rwLock sync.RWMutex
}

func NewBuilderPubkeyCache(builders BuilderRegistry) (*BuilderPubkeyCache, error) {
	builderCount, err := builders.BuilderCount()
	if err != nil {
		return nil, err
	}
	pc := &BuilderPubkeyCache{
		parent:             nil,
		trustedParentCount: 0,
		pub2idx:            make(map[BLSPubkey]BuilderIndex),
		idx2pub:            make([]CachedPubkey, 0),
	}
	currentCount := uint64(len(pc.idx2pub))
	for i := currentCount; i < builderCount; i++ {
		idx := BuilderIndex(i)
		v, err := builders.Builder(idx)
		if err != nil {
			return nil, err
		}
		pub, err := v.Pubkey()
		if err != nil {
			return nil, err
		}
		pc.pub2idx[pub] = idx
		pc.idx2pub = append(pc.idx2pub, CachedPubkey{Compressed: pub})
	}
	return pc, nil
}

func EmptyBuilderPubkeyCache() *BuilderPubkeyCache {
	return &BuilderPubkeyCache{
		parent:             nil,
		trustedParentCount: 0,
		pub2idx:            make(map[BLSPubkey]BuilderIndex),
		idx2pub:            make([]CachedPubkey, 0),
	}
}

// Get the pubkey of a builder index.
// Note: this does not mean the builder is part of the current state.
// It merely means that this is a known pubkey for that particular builder
// (could be in a later part of a forked version of the state).
func (pc *BuilderPubkeyCache) Pubkey(index BuilderIndex) (pub *CachedPubkey, ok bool) {
	pc.rwLock.RLock()
	defer pc.rwLock.RUnlock()
	return pc.unsafePubkey(index)
}

func (pc *BuilderPubkeyCache) unsafePubkey(index BuilderIndex) (pub *CachedPubkey, ok bool) {
	if index >= pc.trustedParentCount {
		if index >= pc.trustedParentCount+BuilderIndex(len(pc.idx2pub)) {
			return nil, false
		}
		return &pc.idx2pub[index-pc.trustedParentCount], true
	} else if pc.parent != nil {
		return pc.parent.Pubkey(index)
	} else {
		return nil, false
	}
}

// Get the builder index of a pubkey.
// Note: this does not mean the builder is part of the current state.
// It merely means that this is a known pubkey for that particular builder
// (could be in a later part of a forked version of the state).
func (pc *BuilderPubkeyCache) BuilderIndex(pubkey BLSPubkey) (index BuilderIndex, ok bool) {
	pc.rwLock.RLock()
	defer pc.rwLock.RUnlock()
	return pc.unsafeBuilderIndex(pubkey)
}

func (pc *BuilderPubkeyCache) unsafeBuilderIndex(pubkey BLSPubkey) (index BuilderIndex, ok bool) {
	index, ok = pc.pub2idx[pubkey]
	if !ok && pc.parent != nil {
		return pc.parent.BuilderIndex(pubkey)
	}
	return index, ok
}

// AddBuilder appends the (index, pubkey) pair to the pubkey cache. It returns the same cache if the added entry is not conflicting.
// If it conflicts, the part is inherited, and a forked pubkey cache is returned.
func (pc *BuilderPubkeyCache) AddBuilder(index BuilderIndex, pub BLSPubkey) (*BuilderPubkeyCache, error) {
	existingIndex, indexExists := pc.BuilderIndex(pub)
	existingPubkey, pubkeyExists := pc.Pubkey(index)

	if indexExists {
		if existingIndex != index {
			// conflict detected! Deposit log fork!
			forkedPc := &BuilderPubkeyCache{
				parent: pc,
				// fork out the existing index, only trust the history
				trustedParentCount: existingIndex,
				pub2idx:            make(map[BLSPubkey]BuilderIndex),
				idx2pub:            make([]CachedPubkey, 0),
			}
			// Do not have to unlock this cache (parent of forkedPc) early, as the forkedPc is guaranteed to handle it.
			return forkedPc.AddBuilder(index, pub)
		}
		if pubkeyExists {
			if existingPubkey.Compressed != pub {
				// conflict detected! Deposit log fork!
				forkedPc := &BuilderPubkeyCache{
					parent: pc,
					// fork out the existing index, only trust the history
					trustedParentCount: index,
					pub2idx:            make(map[BLSPubkey]BuilderIndex),
					idx2pub:            make([]CachedPubkey, 0),
				}
				// Do not have to unlock this cache (parent of forkedPc) early, as the forkedPc is guaranteed to handle it.
				return forkedPc.AddBuilder(index, pub)
			}
		}
		// append is no-op, builder already exists
		return pc, nil
	}
	if pubkeyExists {
		if existingPubkey.Compressed != pub {
			// conflict detected! Deposit log fork!
			forkedPc := &BuilderPubkeyCache{
				parent: pc,
				// fork out the existing index, only trust the history
				trustedParentCount: index,
				pub2idx:            make(map[BLSPubkey]BuilderIndex),
				idx2pub:            make([]CachedPubkey, 0),
			}
			// Do not have to unlock this cache (parent of forkedPc) early, as the forkedPc is guaranteed to handle it.
			return forkedPc.AddBuilder(index, pub)
		}
	}
	pc.rwLock.Lock()
	defer pc.rwLock.Unlock()
	if expected := pc.trustedParentCount + BuilderIndex(len(pc.idx2pub)); index != expected {
		// index is unknown, but too far ahead of cache; in between indices are missing.
		return nil, fmt.Errorf("AddBuilder is incorrect, missing earlier index. got: (%d, %x), but currently expecting %d next", index, pub, expected)
	}
	pc.idx2pub = append(pc.idx2pub, CachedPubkey{Compressed: pub})
	pc.pub2idx[pub] = index
	return pc, nil
}
