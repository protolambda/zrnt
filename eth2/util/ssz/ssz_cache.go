package ssz

import (
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/hash"
	"github.com/protolambda/zrnt/eth2/util/math"
	"reflect"
)

type SSZSerializeCacheProvider interface {
	GetSerializeCache() *SerializeCache
}


type SerializeCache struct {

	Serialized []byte

	// inverse dirty flag. False = cache needs to be filled/refreshed. True = cache is ready to use
	Cached bool
}

type SerializationCacher struct {

	SerializeCache *SerializeCache
}

// returns itself, used to recognize caches from general interfaces.
// Can be inherited to provide cache through embedding.
func (c *SerializationCacher) GetSerializeCache() *SerializeCache {
	// lazy initialize cache
	if c.SerializeCache == nil {
		c.SerializeCache = new(SerializeCache)
	}
	return c.SerializeCache
}


type SSZTreeRootCacheProvider interface {
	GetTreeRootCache() *TreeRootCache
}

type TreeRootCache struct {

	Root beacon.Root

	// inverse dirty flag. False = cache needs to be filled/refreshed. True = cache is ready to use
	Cached bool
}

type TreeRootCacher struct {

	TreeRootCache *TreeRootCache
}

// returns itself, used to recognize caches from general interfaces.
// Can be inherited to provide cache through embedding.
func (c *TreeRootCacher) GetTreeRootCache() *TreeRootCache {
	// lazy initialize cache
	if c.TreeRootCache == nil {
		c.TreeRootCache = new(TreeRootCache)
	}
	return c.TreeRootCache
}

// composition of both cache types
type SSZCaching struct {
	SerializeCache
	TreeRootCacher
}


type SSZCompoundCache struct {
	WorkSheet [][32]byte
	// keep track of what changed in a mapping with the same structure as a worksheet.
	// false = dirty, true = cached
	CacheSheet []bool
	// size of each element. Use 32 if elements are roots.
	ElemSize uint64
	// how many elements there are in the cache
	ElemCount uint64
}

func (cache *SSZCompoundCache) elemIndexToChunkIndex(index uint64) uint64 {
	return (index * cache.ElemSize) >> 5
}

func (cache *SSZCompoundCache) CalcElemSize(vt reflect.Type) {
	switch vt.Kind() {
	case reflect.Uint8:
		cache.ElemSize = 1
	case reflect.Uint32:
		cache.ElemSize = 4
	case reflect.Uint64:
		cache.ElemSize = 8
	case reflect.Bool:
		cache.ElemSize = 1
	case reflect.Slice:
		cache.ElemSize = 32 // length is mixed in with root, but hashed again
	case reflect.Array:
		cache.ElemSize = 32
	case reflect.Struct:
		cache.ElemSize = 32
	default:
		panic("SSZ hasEmbedSize: unsupported value type: " + vt.String())
	}
	div := 32 / cache.ElemSize
	if div * cache.ElemSize != 32 {
		panic("element size is not a factor of 32")
	}
}

// marks an element (by its element index) as changed.
// Call SetLength first if you mark something out of bounds.
func (cache *SSZCompoundCache) SetChanged(index uint64) {
	if cache.ElemSize == 0 {
		panic("ssz cache element size is not initialized with item type")
	}
	if cache.ElemSize > 32 {
		panic("elements cannot be larger than a chunk! (hash them to embed them in the tree)")
	} else if cache.ElemSize < 32 {
		chunkIndex := cache.elemIndexToChunkIndex(index)
		if chunkIndex >= cache.ElemCount {
			panic("changing out of SSZ cache bounds")
		}
		for chunkIndex != 0 {
			cache.CacheSheet[chunkIndex] = false
			chunkIndex >>= 1
		}
	} else {
		if index >= cache.ElemCount {
			panic("changing out of SSZ cache bounds")
		}
		for index != 0 {
			cache.CacheSheet[index] = false
			index >>= 1
		}
	}
}

// Changes the size of the cache, given an element count to facilitate
func (cache *SSZCompoundCache) SetLength(count uint64) {
	power2 := math.NextPowerOfTwo(count)
	thisLen := power2<<1
	prevLen := uint64(len(cache.WorkSheet))
	if prevLen < thisLen {
		// our merkle worksheet, carries in-between hashing work, for re-use later on
		thisWork := make([][32]byte, thisLen)
		thisCached := make([]bool, thisLen)
		// previous work is smaller, we have to extend our merkle worksheet, just 2x difference would like this:
		// old: bbccccddddddddeeeeeeeeeeeeeeee                                  (lowercase is existing trie nodes)
		// how:   bb cccc     dddddddd        eeeeeeeeeeeeeeee
		// new: AAbbBBccccCCCCddddddddDDDDDDDDeeeeeeeeeeeeeeeeEEEEEEEEEEEEEEEE  (uppercase is new trie nodes)
		prevPower2 := math.NextPowerOfTwo(prevLen) >> 1
		diffPower2 := power2 - prevPower2
		for i := uint64(1); i <= prevPower2; i++ {
			srcS := (1 << i)-2
			srcE := (2 << i)-2
			dstS := ((1 << i) << diffPower2)-2
			dstE := ((2 << i) << diffPower2)-2
			copy(thisWork[dstS:dstE], cache.WorkSheet[srcS:srcE])
		}
		cache.WorkSheet = thisWork
		cache.CacheSheet = thisCached
	} else if prevLen > thisLen {
		// our merkle worksheet, carries in-between hashing work, for re-use later on
		thisWork := make([][32]byte, thisLen)
		// previous work is bigger, we can shrink our merkle worksheet, just 2x difference would like this:
		// old: bbccccddddddddeeeeeeeeeeeeeeee          (lowercase is existing trie nodes)
		// how:   cc  dddd    eeeeeeee
		// new: ccddddeeeeeeee                          (uppercase is new trie nodes)
		// effectively the reverse of the above extension of the merkle worksheet
		prevPower2 := math.NextPowerOfTwo(prevLen) >> 1
		diffPower2 := prevPower2 - power2
		for i := uint64(1); i <= power2; i++ {
			srcS := ((1 << i) << diffPower2)-2
			srcE := ((2 << i) << diffPower2)-2
			dstS := (1 << i)-2
			dstE := (2 << i)-2
			copy(thisWork[dstS:dstE], cache.WorkSheet[srcS:srcE])
		}
		cache.WorkSheet = thisWork
		cache.CacheSheet = make([]bool, thisLen, thisLen)
	} else {
		// ideal situation, merkle worksheet sizes match, we don't have to do anything
	}
	cache.ElemCount = count
}

type SSZIndexedSerializer func(dst []byte, index uint64)

func (cache *SSZCompoundCache) UpdateAndMerkleize(serializeIndex SSZIndexedSerializer) [32]byte {
	if cache.ElemCount == 0 {
		return [32]byte{}
	}
	if cache.ElemCount == 1 {
		return cache.WorkSheet[1]
	}
	power2 := math.NextPowerOfTwo(cache.ElemCount)
	for i := int64(power2) - 1; i > 0; i-- {
		// if both source nodes are cached, then their combination does not have to be re-computed
		a := i << 1
		b := a + 1
		cache.CacheSheet[i] = cache.CacheSheet[a] && cache.CacheSheet[b]
	}
	// Now, for the changed elements, start
	if cache.ElemSize > 32 {
		panic("elements cannot be larger than a chunk! (hash them to embed them in the tree)")
	} else if cache.ElemSize < 32 {
		// one chunk contains multiple elements (aligned to 32 byte chunk)]
		chunkIndex := power2
		for i := uint64(0); i < cache.ElemCount; {
			if cache.CacheSheet[chunkIndex] {
				serializeIndex(cache.WorkSheet[chunkIndex][:], i)
			}
			for j := uint64(0); j < 32; j += cache.ElemSize {
				serializeIndex(cache.WorkSheet[chunkIndex][j:j+cache.ElemSize], i)
				i++
			}
			chunkIndex++
		}
	} else {
		// one element per chunk
		chunkIndex := power2
		for i := uint64(0); i < cache.ElemCount; i++ {
			if cache.CacheSheet[chunkIndex] {
				serializeIndex(cache.WorkSheet[chunkIndex][:], i)
			}
			chunkIndex++
		}
	}
	for i := int64(power2) - 1; i > 0; i-- {
		a := i << 1
		b := a + 1
		if !cache.CacheSheet[i] {
			cache.WorkSheet[i] = hash.Hash(append(cache.WorkSheet[a][:], cache.WorkSheet[b][:]...))
		}
	}
	return cache.WorkSheet[1]
}
