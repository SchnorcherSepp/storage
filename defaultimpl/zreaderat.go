package impl

import (
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"io"
)

var _ interf.ReaderAt = (*_ZeroReaderAt)(nil)

type _ZeroReaderAt struct {
	// nope
}

// NewZeroReaderAt is a dummy ReaderAt with no data.
func NewZeroReaderAt() interf.ReaderAt {
	return new(_ZeroReaderAt)
}

//--------------------------------------------------------------------------------------------------------------------//

func (r *_ZeroReaderAt) ReadAt(_ []byte, _ int64) (n int, err error) {
	return 0, io.EOF
}

func (r *_ZeroReaderAt) Close() error {
	return nil
}

func (r *_ZeroReaderAt) Stat() map[string]uint64 {
	return make(map[string]uint64)
}
