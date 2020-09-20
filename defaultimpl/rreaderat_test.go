package impl_test

import (
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	"io"
	"testing"
)

func Test_RamReaderAt(t *testing.T) {
	// init test
	data := []byte{'a', 'b', 'c', 'd', 'e', 'f'}
	r := impl.NewRamReaderAt(data)

	// test
	buf := make([]byte, 3)
	if n, err := r.ReadAt(buf, 0); n != 3 || err != nil || string(buf[:n]) != "abc" {
		t.Fatalf("fail: n=%d, e='%v', s='%s'", n, err, string(buf[:n]))
	}
	if n, err := r.ReadAt(buf, 1); n != 3 || err != nil || string(buf[:n]) != "bcd" {
		t.Fatalf("fail: n=%d, e='%v', s='%s'", n, err, string(buf[:n]))
	}
	if n, err := r.ReadAt(buf, 2); n != 3 || err != nil || string(buf[:n]) != "cde" {
		t.Fatalf("fail: n=%d, e='%v', s='%s'", n, err, string(buf[:n]))
	}
	if n, err := r.ReadAt(buf, 3); n != 3 || err != nil || string(buf[:n]) != "def" {
		t.Fatalf("fail: n=%d, e='%v', s='%s'", n, err, string(buf[:n]))
	}
	if n, err := r.ReadAt(buf, 4); n != 2 || err != io.EOF || string(buf[:n]) != "ef" {
		t.Fatalf("fail: n=%d, e='%v', s='%s'", n, err, string(buf[:n]))
	}
	if n, err := r.ReadAt(buf, 5); n != 1 || err != io.EOF || string(buf[:n]) != "f" {
		t.Fatalf("fail: n=%d, e='%v', s='%s'", n, err, string(buf[:n]))
	}
	if n, err := r.ReadAt(buf, 6); n != 0 || err != io.EOF || string(buf[:n]) != "" {
		t.Fatalf("fail: n=%d, e='%v', s='%s'", n, err, string(buf[:n]))
	}
	if n, err := r.ReadAt(buf, 7); n != 0 || err != io.EOF || string(buf[:n]) != "" {
		t.Fatalf("fail: n=%d, e='%v', s='%s'", n, err, string(buf[:n]))
	}

	// close test
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	// test nil request
	if n, err := r.ReadAt(nil, 0); n != 0 || err != nil { // no request, no data
		t.Fatalf("fail: n=%d, e='%v'", n, err)
	}

	// test nil data
	r = impl.NewRamReaderAt(nil)
	if n, err := r.ReadAt(buf, 0); n != 0 || err != io.EOF || string(buf[:n]) != "" {
		t.Fatalf("fail: n=%d, e='%v', s='%s'", n, err, string(buf[:n]))
	}
}
