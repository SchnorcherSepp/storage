package impl_test

import (
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"io"
	"sync"
	"testing"
)

func Test_SubReaderAt_ReadAt_f0(t *testing.T) {
	// build test Service
	cache := impl.NewCache(1)
	s := impl.NewRamService(cache, false)
	if err := impl.InitDemo(s); err != nil {
		t.Fatal(err)
	}

	// select files for tests
	f, err := s.Files().ByName("special-file-0-0.000000.dat")
	if err != nil {
		t.Fatal(err)
	}

	// test
	buf := make([]byte, 3)
	ra, err := impl.NewSubReaderAt(f, s, cache, false, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	rb, err := impl.NewSubReaderAt(f, s, cache, false, 0, 15)
	if err != nil {
		t.Fatal(err)
	}
	rc, err := impl.NewSubReaderAt(f, s, cache, false, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	//-----------------------------------------------
	for _, r := range []interf.ReaderAt{ra, rb, rc} {
		if n, e := r.ReadAt(buf, 0); e != io.EOF || n != 0 {
			t.Fatalf("test error: n=%d, e=%v", n, e)
		}
		if n, e := r.ReadAt(buf, 1); e != io.EOF || n != 0 {
			t.Fatalf("test error: n=%d, e=%v", n, e)
		}
	}
}

func Test_SubReaderAt_ReadAt_f1(t *testing.T) {
	// build test Service
	cache := impl.NewCache(1)
	s := impl.NewRamService(cache, false)
	if err := impl.InitDemo(s); err != nil {
		t.Fatal(err)
	}

	// select files for tests
	f, err := s.Files().ByName("small-test-file-9.dat")
	if err != nil {
		t.Fatal(err)
	}

	// test A: normal limit
	// "small-test-file-9.dat"
	buf := make([]byte, 100)
	ra, err := impl.NewSubReaderAt(f, s, cache, false, 0, 15)
	if err != nil {
		t.Fatal(err)
	}
	//-----------------------------------------------
	{
		r := ra
		if n, e := r.ReadAt(buf, 0); e != io.EOF || n != 15 || string(buf[:n]) != "small-test-file" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 1); e != io.EOF || n != 14 || string(buf[:n]) != "mall-test-file" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 14); e != io.EOF || n != 1 || string(buf[:n]) != "e" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 15); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 16); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		buf = make([]byte, 3)
		if n, e := r.ReadAt(buf, 0); e != nil || n != 3 || string(buf[:n]) != "sma" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 1); e != nil || n != 3 || string(buf[:n]) != "mal" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 14); e != io.EOF || n != 1 || string(buf[:n]) != "e" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 15); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 30); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
	}

	// test B: normal limit with offset
	// "small-test-file-9.dat"
	buf = make([]byte, 100)
	rb, err := impl.NewSubReaderAt(f, s, cache, false, 5, 15)
	if err != nil {
		t.Fatal(err)
	}
	//-----------------------------------------------
	{
		r := rb
		if n, e := r.ReadAt(buf, 0); e != io.EOF || n != 15 || string(buf[:n]) != "-test-file-9.da" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 1); e != io.EOF || n != 14 || string(buf[:n]) != "test-file-9.da" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 14); e != io.EOF || n != 1 || string(buf[:n]) != "a" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 15); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 16); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		buf = make([]byte, 3)
		if n, e := r.ReadAt(buf, 0); e != nil || n != 3 || string(buf[:n]) != "-te" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 1); e != nil || n != 3 || string(buf[:n]) != "tes" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 14); e != io.EOF || n != 1 || string(buf[:n]) != "a" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
	}

	// test C: limit == filesize
	// "small-test-file-9.dat"
	buf = make([]byte, 100)
	rc, err := impl.NewSubReaderAt(f, s, cache, false, 5, 16)
	if err != nil {
		t.Fatal(err)
	}
	//-----------------------------------------------
	{
		r := rc
		if n, e := r.ReadAt(buf, 0); e != io.EOF || n != 16 || string(buf[:n]) != "-test-file-9.dat" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 1); e != io.EOF || n != 15 || string(buf[:n]) != "test-file-9.dat" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 14); e != io.EOF || n != 2 || string(buf[:n]) != "at" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 15); e != io.EOF || n != 1 || string(buf[:n]) != "t" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 16); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 17); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		buf = make([]byte, 3)
		if n, e := r.ReadAt(buf, 12); e != nil || n != 3 || string(buf[:n]) != ".da" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 13); e != nil || n != 3 || string(buf[:n]) != "dat" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 14); e != io.EOF || n != 2 || string(buf[:n]) != "at" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 15); e != io.EOF || n != 1 || string(buf[:n]) != "t" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 16); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 17); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
	}

	// test D: limit over file size
	// "small-test-file-9.dat"
	buf = make([]byte, 3)
	rd, err := impl.NewSubReaderAt(f, s, cache, false, 17, 33)
	if err != nil {
		t.Fatal(err)
	}
	//-----------------------------------------------
	{
		r := rd
		if n, e := r.ReadAt(buf, 0); e != nil || n != 3 || string(buf[:n]) != ".da" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 1); e != nil || n != 3 || string(buf[:n]) != "dat" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 2); e != io.EOF || n != 2 || string(buf[:n]) != "at" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 3); e != io.EOF || n != 1 || string(buf[:n]) != "t" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 4); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
		if n, e := r.ReadAt(buf, 5); e != io.EOF || n != 0 || string(buf[:n]) != "" {
			t.Fatalf("test error: n=%d, e=%v, s=%s", n, e, string(buf[:n]))
		}
	}
}

//--------------------------------------------------------------------------------------------------------------------//

func TestRace_SubReaderAt(t *testing.T) {
	f, s, _ := initTestFileAndTestService(t)

	r, err := impl.NewSubReaderAt(f, s, nil, false, 1, 999)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(5)
	for n := 0; n < 5; n++ {
		loop := n
		go func() {
			//------------------------------
			for i := 0; i < 999; i++ {
				n, err1 := r.ReadAt(make([]byte, 1), int64(i))
				err2 := r.Close()
				r.Stat()
				if err1 != nil || err2 != nil || n != 1 {
					t.Errorf("e1=%v, e2=%v, loop=%d, index=%d", err1, err2, loop, i)
					break
				}
			}
			//------------------------------
			wg.Done()
		}()
	}
	wg.Wait()
}
