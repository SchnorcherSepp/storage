package impl

import (
	"crypto/md5"
	"errors"
	"fmt"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"io"
	"sync"
)

// interface check: interf.ReaderAt
var _ interf.ReaderAt = (*_MReaderAt)(nil)

// @see interf.ReaderService
// @see interf.ReaderAt
//
// MultiReaderAt allow random read access to a series of files identified by the file ids.
// All files except the last file must be the same size.
// In addition, this method behaves like ReaderAt.
type _MReaderAt struct {
	readers     []interf.ReaderAt
	files       []interf.File
	fileSize    int64
	mux         *sync.RWMutex
	stat        *_ReaderStat
	multiFileId string
}

// NewMultiReader combine two or more ReaderAt and behave like a normal ReaderAT for a single file.
// All files except the last file must be the same size.
// There must be at least two or more files!
func NewMultiReaderAt(files []interf.File, service interf.ReaderService, cache interf.Cache, debugLvl uint8) (interf.ReaderAt, error) {
	// ReaderAt statistic
	stat := &_ReaderStat{
		debugLvl:    debugLvl,       // enable debug logging [0, 1, 2] (level: high=2)
		packageName: "[MULTI] impl", // text for debug logging
	}

	// at least one file
	if files == nil || len(files) <= 1 || service == nil {
		return nil, errors.New("can't create new NewMultiReaderAt with file=nil, len(file)=1 or service=nil")
	}

	// all files (except the last) must have the same size
	fileSize := files[0].Size()
	for i, v := range files {
		if v.Size() == 0 || v.Size() != fileSize && i != len(files)-1 {
			return nil, errors.New("MultiReaderAt can't combine files of different sizes or empty files")
		}
	}

	// create all inner ReaderAt
	h := md5.New()
	readers := make([]interf.ReaderAt, len(files))
	for i, f := range files {
		r, err := NewReaderAt(f, service, cache, debugLvl)
		if err != nil {
			// error from NewReaderAt()
			return nil, err
		}
		readers[i] = r
		h.Write([]byte(f.Id()))
	}
	multiFileId := fmt.Sprintf("%x", h.Sum(nil))

	// return
	stat.RAtNew(multiFileId, cache != nil) // DEBUG
	return &_MReaderAt{
		readers:     readers,
		files:       files,
		fileSize:    fileSize,
		mux:         new(sync.RWMutex),
		stat:        stat,
		multiFileId: multiFileId,
	}, nil
}

// @see interf.ReaderAt
func (r *_MReaderAt) Close() error {
	r.mux.Lock() // LOCK
	defer r.mux.Unlock()

	r.stat.RAtClosing(r.multiFileId) // DEBUG
	if r.readers != nil {
		for i, inner := range r.readers {
			if inner != nil {
				r.stat.RAtClose(r.multiFileId, i, true) // DEBUG
				_ = inner.Close()
			}
		}
	}

	r.stat.PrintStatAfterClose(r.multiFileId) // DEBUG
	return nil
}

// @see interf.ReaderAt
func (r *_MReaderAt) ReadAt(p []byte, off int64) (int, error) {
	r.mux.RLock() // READ LOCK
	defer r.mux.RUnlock()

	// check fast return
	if len(p) == 0 {
		return 0, nil // read nothing -> return nothing
	}

	// calc file and offset
	fileOff := off % r.fileSize
	fileNo := int((off - fileOff) / r.fileSize)

	// request
	r.stat.RAtReq(r.multiFileId, off, len(p), uint64(fileNo), int(fileOff)) // DEBUG

	var read int
	var err error
	for {
		var n int

		// file No must exist
		if fileNo < len(r.readers) {
			// delegate to inner ReaderAt
			n, err = r.readers[fileNo].ReadAt(p[read:], fileOff)
		} else {
			// no file found
			n, err = 0, io.EOF
		}

		// update vars
		fileNo++    // next file
		fileOff = 0 // fileOff is 0 after first read
		read += n   // update read n

		// exit
		if err != nil && err != io.EOF {
			// serious error! (no EOF)
			break
		} else {
			if n == 0 {
				// no data! EOF?
				break
			}
			if read == len(p) {
				// success: all we need
				// fix: ignore EOF (buffer is full)
				err = nil
				break
			}
		}
	}

	// fix EOF
	if len(p) > 0 && read <= 0 && err == nil {
		err = io.EOF
	}

	// return
	r.stat.RAtRet(r.multiFileId, off, len(p), read, err) // DEBUG
	return read, err
}

// @see interf.ReaderAt
//
// Stat returns the number of times internal processes have been run since initialization.
// This method is relevant for testing and debugging purposes.
// The KEY is the internal process, the VALUE is the count.
func (r *_MReaderAt) Stat() map[string]uint64 {
	r.mux.RLock() // READ LOCK
	defer r.mux.RUnlock()

	// summary
	ret := make(map[string]uint64)

	// _MReaderAt stats
	for k, v := range r.stat.Stat() {
		if v > 0 {
			ret["[MULTI] "+k] = v
		}
	}

	// inner stats
	for i, inner := range r.readers {
		for k, v := range inner.Stat() {
			if v > 0 {
				ret[fmt.Sprintf("[%d] %s", i, k)] = v
			}
		}
	}

	return ret
}
