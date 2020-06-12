package impl

import (
	"encoding/binary"
	"github.com/SchnorcherSepp/storage/interfaces"
	"github.com/coocood/freecache"
	"github.com/oxtoacart/bpool"
	"runtime/debug"
)

// interface check: interf.Cache
var _ interf.Cache = (*_Cache)(nil)

// @see interf.Cache
//
// Cache stores sectors (data blocks of a file) for a performant random read access. (@see interf.ReadAt)
// The cache is always at least 1024 * SectorSize big (~17 MB).
// If possible, there should only be one common large cache (reuse the object in your program).
type _Cache struct {
	cache *freecache.Cache // RAM cache for sectors
	pool  *bpool.BytePool  // buffer pool
}

// NewCache return the default implementation of interf.Cache.
// cacheSizeMB can't be less than 17 (min. 1024 * SectorSize =~ 17 MB).
func NewCache(cacheSizeMB int) interf.Cache {
	// cache min. size
	min := ((1024 * interf.SectorSize) / (1024 * 1024)) + 1
	if cacheSizeMB < min {
		cacheSizeMB = min
	}

	// inti freeCache
	cacheSize := cacheSizeMB * 1024 * 1024
	fCache := freecache.NewCache(cacheSize) // > 17 MB
	debug.SetGCPercent(20)

	return &_Cache{
		cache: fCache,
		pool:  bpool.NewBytePool(300, interf.SectorSize), // ~ 5 MB
	}
}

// @see interf.Cache
//
// Get returns the value or 'not found' error.
// This method doesn't allocate memory when the capacity of buf is greater or equal to value.
func (c *_Cache) Get(fileId string, sector uint64, buf []byte) ([]byte, error) {
	key := c.calcCacheKey(fileId, sector)
	return c.cache.GetWithBuf(key, buf)
}

// @see interf.Cache
//
// Set stores the value in the cache.
// Old data can be deleted if the cache is full.
// The value expires after interf.CacheExpireSeconds.
func (c *_Cache) Set(fileId string, sector uint64, data []byte) error {
	key := c.calcCacheKey(fileId, sector)
	return c.cache.Set(key, data, interf.CacheExpireSeconds)
}

// @see interf.Cache
//
// Pool returns a byte pool. This means that the small byte buffers can be reused and the allocation is reduced.
// The Pool contain 300 buffer with the size of interf.SectorSize.
//
// Example of use:
//   buf := c.Pool().Get()
//   defer c.Pool().Put(buf)
func (c *_Cache) Pool() *bpool.BytePool {
	return c.pool
}

//-----  HELPER  -----------------------------------------------------------------------------------------------------//

// calcCacheKey converts fileId and a sector into a byte key for freeCache.
func (c *_Cache) calcCacheKey(fileId string, sector uint64) []byte {
	var bKey [8]byte
	binary.LittleEndian.PutUint64(bKey[:], sector)
	return append(bKey[:], []byte(fileId)...)
}
