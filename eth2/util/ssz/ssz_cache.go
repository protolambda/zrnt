package ssz

import "github.com/protolambda/zrnt/eth2/beacon"

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


type SSZCompoundCache []*SSZCaching


