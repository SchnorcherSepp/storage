package interf

import "github.com/oxtoacart/bpool"

// Cache stores sectors (data blocks of a file) for a performant random read access. (@see interf.ReadAt)
// The cache is always at least 1024 * SectorSize big (~17 MB).
// If possible, there should only be one common large cache (reuse the object in your program).
type Cache interface {

	// Get returns the value or 'not found' error.
	// This method doesn't allocate memory when the capacity of buf is greater or equal to value.
	Get(fileId string, sector uint64, buf []byte) ([]byte, error)

	// Set stores the value in the cache.
	// Old data can be deleted if the cache is full.
	// The value expires after interf.CacheExpireSeconds.
	Set(fileId string, sector uint64, data []byte) error

	// Pool returns a byte pool. This means that the small byte buffers can be reused and the allocation is reduced.
	// The Pool contain 300 buffer with the size of interf.SectorSize.
	//
	// Example of use:
	//   buf := c.Pool().Get()
	//   defer c.Pool().Put(buf)
	Pool() *bpool.BytePool

	// Size returns the max. capacity of this cache in bytes.
	Size() int64
}
