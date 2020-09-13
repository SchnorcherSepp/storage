package impl

import (
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"io"
)

// interface check: interf.ReaderAt
var _ interf.ReaderAt = (*_SubReaderAt)(nil)

// @see interf.ReaderAt
//
// SubReaderAt is a wrapper for impl.ReaderAt (@see impl.NewReaderAt).
// This implementation enables the ReaderAt methods to a part of a file.
type _SubReaderAt struct {
	inner interf.ReaderAt
	off   int64
	n     int64
}

// NewSubReaderAt creates a new interf.ReaderAt object for random read access to the file.
// This implementation enables the ReaderAt methods to a part of a file.
// No connections are made before the first call of ReadAt().
// Is cache = nil, the cache is disabled.
// The offset off is the part start point and n the part size.
func NewSubReaderAt(file interf.File, service interf.ReaderService, cache interf.Cache, debugLog bool, off, n int64) (interf.ReaderAt, error) {

	// get normal ReaderAt
	rAt, err := NewReaderAt(file, service, cache, debugLog)

	// build SubReaderAt
	return &_SubReaderAt{
		inner: rAt,
		off:   off,
		n:     n,
	}, err
}

// @see interf.ReaderAt
func (r *_SubReaderAt) Close() error {
	return r.inner.Close()
}

// @see interf.ReaderAt
func (r *_SubReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	// inner call
	n, err = r.inner.ReadAt(p, r.off+off)

	// check n (enforce limit)
	startP := r.off + off
	endPos := startP + int64(n)
	maxEPos := r.off + r.n
	if endPos > maxEPos {
		// update n
		endPos = maxEPos
		n = int(endPos - startP)
		// enforce min n = 0
		if n < 0 {
			n = 0
		}
		// fix EOF for limit: err is nil AND buffer is NOT full!
		if len(p) > n && err == nil {
			err = io.EOF
		}
	}

	// fix EOF for no data
	if n == 0 && err == nil {
		err = io.EOF
	}

	// return
	return
}

// @see interf.ReaderAt
//
// Stat returns the number of times internal processes have been run since initialization.
// This method is relevant for testing and debugging purposes.
// The KEY is the internal process, the VALUE is the count.
func (r *_SubReaderAt) Stat() map[string]uint64 {
	return r.inner.Stat()
}
