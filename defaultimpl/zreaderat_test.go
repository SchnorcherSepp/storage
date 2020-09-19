package impl_test

import (
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	"io"
	"testing"
)

func Test_ZeroReaderAt(t *testing.T) {
	r := impl.NewZeroReaderAt()

	// test close
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	// test stat
	if st := r.Stat(); st == nil {
		t.Fatal("error")
	}

	// test read
	if n, err := r.ReadAt(nil, 0); n != 0 || err != io.EOF {
		t.Fatal("error")
	}
	if n, err := r.ReadAt(make([]byte, 0), 0); n != 0 || err != io.EOF {
		t.Fatal("error")
	}
	if n, err := r.ReadAt(make([]byte, 3), 1); n != 0 || err != io.EOF {
		t.Fatal("error")
	}
	if n, err := r.ReadAt(make([]byte, 5), 15); n != 0 || err != io.EOF {
		t.Fatal("error")
	}
}
