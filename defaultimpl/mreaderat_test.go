package impl_test

import (
	"fmt"
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"io"
	"log"
	"math/rand"
	"reflect"
	"sync"
	"testing"
)

func TestNewMultiReaderAt(t *testing.T) {
	_, s, f := initTestFileAndTestService(t)

	// test with invalid file and invalid service
	if _, err := impl.NewMultiReaderAt(nil, s, nil, impl.DebugHigh); err == nil {
		t.Fatal("no error with invalid file")
	}
	if _, err := impl.NewMultiReaderAt(f, nil, nil, impl.DebugHigh); err == nil {
		t.Fatal("no error with invalid file")
	}

	// different file size
	_, err := impl.NewMultiReaderAt([]interf.File{f[1], f[0], f[1]}, s, nil, impl.DebugHigh)
	if err == nil {
		t.Fatal("no error")
	}

	// only one file
	_, err = impl.NewMultiReaderAt([]interf.File{f[1]}, s, nil, impl.DebugHigh)
	if err == nil {
		t.Fatal("no error")
	}

	// test without cache
	_, err = impl.NewMultiReaderAt(f, s, nil, impl.DebugHigh)
	if err != nil {
		t.Fatal(err)
	}

	// test with cache
	c := impl.NewCache(1)
	_, err = impl.NewMultiReaderAt(f, s, c, impl.DebugHigh)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_MultiReaderAt_ReadAt(t *testing.T) {
	_, s, f := initTestFileAndTestService(t)
	r, err := impl.NewMultiReaderAt(f, s, nil, impl.DebugHigh)
	if err != nil {
		t.Fatal(err)
	}

	// TEST: read first byte from first file
	b := make([]byte, 1)
	if n, err := r.ReadAt(b, 0); n != 1 || err != nil || b[0] != 38 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// TEST: read last byte from first file
	if n, err := r.ReadAt(b, 150*1024*1024); n != 1 || err != nil || b[0] != 254 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// TEST: read first byte from last file
	if n, err := r.ReadAt(b, 150*1024*1024+1); n != 1 || err != nil || b[0] != 115 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// TEST: read last byte from last file
	if n, err := r.ReadAt(b, 150*1024*1024+21); n != 1 || err != nil || b[0] != 116 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// TEST: over read
	b = make([]byte, 1)
	if n, err := r.ReadAt(b, 150*1024*1024+21+1); n != 0 || err != io.EOF || b[0] != 0 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// TEST: read first AND last file
	b = make([]byte, 2)
	if n, err := r.ReadAt(b, 150*1024*1024); n != 2 || err != nil || b[0] != 254 || b[1] != 115 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// PRINT STATS
	log.Printf("%#v", r.Stat())
}

func Test_MReaderAt_ReadAt__RandomRead(t *testing.T) {
	_, service, _ := initTestFileAndTestService(t)

	files := make([]interf.File, 0)
	data := make([]byte, 0)

	for i := 999; i > 98; i-- {
		name := fmt.Sprintf("small-test-file-%d.dat", i)
		f, err := service.Files().ByName(name)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
		data = append(data, []byte(name)...)
	}

	// TEST
	for _, cache := range []interf.Cache{nil, impl.NewCache(1)} {
		// stuff
		rnd := rand.New(rand.NewSource(1234567890))
		buf := make([]byte, 128)

		// multi reader
		r, err := impl.NewMultiReaderAt(files, service, cache, impl.DebugOff)
		if err != nil {
			t.Fatal(err)
		}

		// random read tests
		for i := 0; i < 50000; i++ {
			// read
			off := rnd.Intn(len(data) * 2)
			n, err := r.ReadAt(buf, int64(off))

			// get data sample
			var testData []byte
			if off > len(data) {
				testData = []byte{}
			} else {
				to := off + len(buf)
				if to > len(data) {
					to = len(data)
				}
				testData = data[off:to]
			}
			var testErr error
			if len(testData) == 0 || len(testData) != len(buf) {
				testErr = io.EOF
			}

			// check
			if err != testErr || len(testData) != n || !reflect.DeepEqual(buf[:n], testData) {
				t.Errorf("round=%d, off=%d/%d, data=%d, testdata=%d, err=%v, testErr=%v", i, off, len(data), len(buf[:n]), len(testData), err, testErr)
			}
		}
	}
}

//--------------------------------------------------------------------------------------------------------------------//

func TestRace_MultiReaderAt(t *testing.T) {
	_, s, f := initTestFileAndTestService(t)

	r, err := impl.NewMultiReaderAt(f, s, nil, impl.DebugOff) // test without cache for more inner code tests
	if err != nil {
		t.Fatal(err)
	}

	const rounds = 5
	var wg sync.WaitGroup
	wg.Add(rounds)
	for round := 0; round < rounds; round++ {
		go func() {
			//------------------------------
			for off := int64(0); off < 1000; off++ {
				b := make([]byte, 1)
				n, err1 := r.ReadAt(b, off)
				err2 := r.Close()
				r.Stat()
				if err1 != nil || err2 != nil || n != 1 {
					t.Errorf("err1=%v, err2=%v, n=%d, off=%d", err1, err2, n, off)
					break
				}
			}
			//------------------------------
			wg.Done()
		}()
	}
	wg.Wait()
}
