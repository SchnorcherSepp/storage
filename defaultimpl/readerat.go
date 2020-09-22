package impl

import (
	"errors"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"github.com/oxtoacart/bpool"
	"io"
	"math"
	"sort"
	"sync"
	"time"
)

// interface check: interf.ReaderAt
var _ interf.ReaderAt = (*_ReaderAt)(nil)

// @see interf.ReaderAt
//
// ReaderAt allow random read access to a file identified by the file id.
// A cache must be used internally for random read access.
// It may also be necessary to open several internal connections to the storage.
type _ReaderAt struct {
	mux   *sync.Mutex  // protect 'inner'
	inner []*_Reader   // open connections to the file (backbone)
	stat  *_ReaderStat // collects statistical data about internal processes

	file    interf.File          // for new connections
	service interf.ReaderService // storage Service (for new connections)
	cache   interf.Cache         // for caching sectors, can be nil !
	pool    *bpool.BytePool      // the byte pool avoids allocating memory

}

// NewReaderAt creates a new interf.ReaderAt object for random read access to the file.
// No connections are made before the first call of ReadAt().
// Is cache = nil, the cache is disabled.
func NewReaderAt(file interf.File, service interf.ReaderService, cache interf.Cache, debugLog bool) (interf.ReaderAt, error) {
	// check input
	// the cache can be nil!
	if file == nil || service == nil {
		return nil, errors.New("can't create new ReaderAt with file=nil or service=nil")
	}

	// ReaderAt statistic
	stat := &_ReaderStat{
		logging:     debugLog, // enable debug logging
		packageName: "impl",   // text for debug logging
	}

	// use byte pool from cache
	// or create a small pool (cache == nil)
	var pool *bpool.BytePool
	if cache != nil {
		pool = cache.Pool()
	} else {
		pool = bpool.NewBytePool(25, interf.SectorSize)
	}

	// return new ReaderAt
	stat.RAtNew(file.Id(), cache != nil) // DEBUG
	return &_ReaderAt{
		mux:   new(sync.Mutex),
		inner: make([]*_Reader, interf.MaxReadersPerFile),
		stat:  stat,

		file:    file,
		service: service,
		cache:   cache,
		pool:    pool,
	}, nil
}

// @see interf.ReaderAt
func (r *_ReaderAt) Close() error {
	r.mux.Lock() // LOCK
	defer r.mux.Unlock()

	r.stat.RAtClosing(r.file.Id()) // DEBUG
	if r.inner != nil {
		for i, v := range r.inner {
			if v != nil {
				r.stat.RAtClose(r.file.Id(), i, v.c != nil) // DEBUG
				_ = v.Close()
				r.inner[i] = nil
			}
		}
	}

	r.stat.PrintStatAfterClose(r.file.Id()) // DEBUG
	return nil
}

// @see interf.ReaderAt
func (r *_ReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if len(p) == 0 {
		return 0, nil // read nothing -> return nothing
	}

	// buffer from pool
	buf := r.pool.Get()
	defer r.pool.Put(buf)

	// read sectors
	sector, innerOff := r.calcSector(off)
	read := 0

	r.stat.RAtReq(r.file.Id(), off, len(p), sector, innerOff) // DEBUG
	for {
		// read sector
		b, err := r.getSector(buf, sector) // thread-safe

		// cut inner offset
		if len(b) < innerOff {
			b = b[len(b):] // nothing left (data are not in this slice! inner offset is to high)
		} else {
			b = b[innerOff:]
		}

		// copy to return buffer
		n := copy(p[read:], b) // copy sector bytes to return buffer

		// update vars
		sector++     // next sector
		innerOff = 0 // innerOff is 0 after first read
		read += n    // update read n

		// exit
		if n == 0 || err != nil || read == len(p) {
			// exit loop, but ...
			// ... fix wrong EOF
			if err == io.EOF && len(p) == read {
				err = nil // a full buffer is never io.EOF
			}
			// write debug and return
			r.stat.RAtRet(r.file.Id(), off, len(p), read, err) // DEBUG
			return read, err
		}
	}
}

// @see interf.ReaderAt
//
// Stat returns the number of times internal processes have been run since initialization.
// This method is relevant for testing and debugging purposes.
// The KEY is the internal process, the VALUE is the count.
func (r *_ReaderAt) Stat() map[string]uint64 {
	return r.stat.Stat()
}

//--------  HELPER  --------------------------------------------------------------------------------------------------//

// getSector returns the requested sector.
// This method doesn't allocate memory when the capacity of buf is greater or equal to value (see SectorSize).
func (r *_ReaderAt) getSector(buf []byte, sector uint64) ([]byte, error) {
	r.mux.Lock() // LOCK
	defer r.mux.Unlock()

	// ask cache
	if r.cache != nil {
		b, err := r.cache.Get(r.file.Id(), sector, buf)
		r.stat.CacheGet(r.file.Id(), sector, len(buf), len(b), err) // DEBUG
		if err == nil {
			return b, nil
		}
	}

	// Get best connection
	c := r.bestConn(sector)
	if c == nil {
		// no reader found, create new one
		var err error
		c, err = r.addConn(sector)
		if err != nil {
			// only if service.Reader() fail
			return buf[:0], err
		}
	}

	// check reader distance (off == reqOff?)
	for c.sector < sector {
		logSector := c.sector
		n, err := c.Read(buf)
		r.stat.RAtSectorSkip(r.file.Id(), logSector, n, err) // DEBUG

		if r.cache != nil && n > 0 && (err == nil || err == io.EOF) {
			errSet := r.cache.Set(r.file.Id(), c.sector-1, buf[:n])        // don't waste VALID data
			r.stat.CacheSet(r.file.Id(), c.sector-1, len(buf[:n]), errSet) // DEBUG
		}

		if err != nil {
			// ERROR but
			// we are not where we wanted to be!
			_ = c.Close()       // error -> close connection
			return buf[:0], err // return zero data! we are not at reqOff!
		}
	}

	// read
	n, err := c.Read(buf)
	if err != nil {
		_ = c.Close() // error -> close connection
	}
	r.stat.RAtSectorRet(r.file.Id(), sector, n, err) // DEBUG

	// cache
	if r.cache != nil && n > 0 && (err == nil || err == io.EOF) {
		errSet := r.cache.Set(r.file.Id(), c.sector-1, buf[:n])
		r.stat.CacheSet(r.file.Id(), c.sector-1, len(buf[:n]), errSet) // DEBUG
	}

	return buf[:n], err
}

// bestConn looks for an open connection that can be reused. Returns nil if no valid connection was found.
// Attention: The returned connection does not have to exactly match the desired sector.
func (r *_ReaderAt) bestConn(sector uint64) *_Reader {
	var bestDist uint64 = math.MaxUint64
	var index = -1 // default: -1 (no connection found)

	// search index of the best connection
	for k, v := range r.inner {
		// skip: no valid connection
		if v == nil || v.c == nil {
			continue
		}
		// skip: reqOff is before the position (can't read back) or too far away
		if sector < v.sector || sector > v.sector+interf.MaxSectorJump {
			continue
		}
		// calc distance
		dist := sector - v.sector
		if dist < bestDist {
			// better connection found
			bestDist = dist
			index = k
		}
		// FAST FIN: there is nothing better than 0!
		if bestDist == 0 {
			break
		}
	}

	// return best connection
	if index >= 0 {
		c := r.inner[index]
		r.stat.RAtBest(r.file.Id(), index, c.sector) // DEBUG
		return c
	} else {
		r.stat.RAtBest(r.file.Id(), index, math.MaxUint64) // DEBUG
		return nil                                         // no connection found
	}
}

// sortByAge sort connection by age.
func (r *_ReaderAt) sortByAge() {

	sort.Slice(r.inner, func(p, q int) bool {
		var rP = r.inner[p]            // connection p
		var rQ = r.inner[q]            // connection q
		var ageP int64 = math.MinInt64 // age for invalid connection p
		var ageQ int64 = math.MinInt64 // age for invalid connection q

		// Set age (only valid connection)
		if rP != nil && rP.c != nil {
			ageP = rP.age
		}
		if rQ != nil && rQ.c != nil {
			ageQ = rQ.age
		}

		return ageP > ageQ
	})
}

// addConn opens a new reader/connection and places it first in the internal list.
// The oldest connection is closed.
func (r *_ReaderAt) addConn(sector uint64) (*_Reader, error) {

	// sort
	r.sortByAge()

	// close last position
	last := len(r.inner) - 1
	if r.inner[last] != nil {
		_ = r.inner[last].Close()
	}

	// clear position one
	for i := len(r.inner) - 1; i > 0; i-- {
		r.inner[i] = r.inner[i-1]
	}
	r.inner[0] = nil

	// create new connection
	inner, err := r.service.Reader(r.file, int64(sector*interf.SectorSize))
	r.stat.RAtAdd(r.file.Id(), sector, err) // DEBUG

	if err != nil {
		// service.Reader() error
		return nil, err

	} else {
		// OK! Set connection and return
		r.inner[0] = newInnerReader(inner, sector)
		return r.inner[0], err
	}
}

// calcSector calculates in which sector the first byte begins with a inner offset.
// A file is divided into sectors that are addressed with the sector number.
// The first sector starts at 0.
func (r *_ReaderAt) calcSector(offset int64) (sector uint64, innerOff int) {
	if offset >= 0 {
		// valid offset -> calc stuff
		innerOff = int(offset % interf.SectorSize)
		sector = uint64(offset-int64(innerOff)) / interf.SectorSize
		return

	} else {
		// invalid offset -> return 0
		return 0, 0
	}
}

// ------------------------------------------------------------------------------------------------------------------ //

// interface check: io.ReadCloser
var _ io.ReadCloser = (*_Reader)(nil)

// _Reader is a ReadCloser that stores the current position and time of the last access.
type _Reader struct {
	c      io.ReadCloser // connection to google drive (can be nil)
	sector uint64        // position (sector number) for next read
	age    int64         // time of last use (unix nano)
}

// newInnerReader initialized a new _Reader. sector is the start sector (offset)
func newInnerReader(c io.ReadCloser, sector uint64) *_Reader {
	return &_Reader{
		c:      c,
		sector: sector,
		age:    time.Now().UnixNano(),
	}
}

// Close the connection. Has no effect after the first call.
// It also invalidates the inner connection with nil.
func (r *_Reader) Close() error {
	if r.c != nil {
		_ = r.c.Close()
		r.c = nil
	}
	return nil
}

// Read reads exactly len(buf) bytes from r into buf. If the given buffer is not exactly the
// sector size, an error is returned. When Read encounters an error or end-of-file condition
// after successfully reading n > 0 bytes, it returns the number of bytes read AND the
// (non-nil) error from the same call. Callers should always process the n > 0 bytes returned
// before considering the error err.
func (r *_Reader) Read(buf []byte) (n int, err error) {
	// check connection
	if r.c == nil {
		return 0, io.ErrClosedPipe
	}
	// check buffer size
	if len(buf) != interf.SectorSize {
		return 0, errors.New("wrong buffer size for reading a sector")
	}

	// read all: leave the loop with full buffer or an error
	for n < interf.SectorSize && err == nil {
		var nn int
		nn, err = r.c.Read(buf[n:])
		n += nn
	}

	// update attributes
	if n > 0 {
		r.age = time.Now().UnixNano()
		r.sector += 1
	}

	// buffer is full, everything is fine
	if n >= interf.SectorSize {
		return n, nil // ignore any errors that may have occurred
	}

	// The buffer is not full AND there must be an error. Otherwise the read loop would not
	// have been left. return error and what we read, this connection is done.
	return
}
