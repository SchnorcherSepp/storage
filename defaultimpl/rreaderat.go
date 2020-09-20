package impl

import (
	"errors"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"io"
)

var _ interf.ReaderAt = (*_RamReaderAt)(nil)

type _RamReaderAt struct {
	data []byte
}

// NewRamReaderAt return a ReaderAt implementation that provides data from the ram ([]byte).
func NewRamReaderAt(data []byte) interf.ReaderAt {
	// check nil
	if data == nil {
		data = make([]byte, 0)
	}
	// return
	return &_RamReaderAt{
		data: data,
	}
}

//--------------------------------------------------------------------------------------------------------------------//

func (r *_RamReaderAt) ReadAt(b []byte, off int64) (n int, err error) {
	// check off
	if off < 0 {
		return 0, errors.New("bytes.Reader.ReadAt: negative offset")
	}
	// no data
	if off >= int64(len(r.data)) {
		return 0, io.EOF
	}
	// copy & return
	n = copy(b, r.data[off:])
	if n < len(b) {
		err = io.EOF
	}
	return
}

func (r *_RamReaderAt) Close() error {
	return nil
}

func (r *_RamReaderAt) Stat() map[string]uint64 {
	return make(map[string]uint64)
}
