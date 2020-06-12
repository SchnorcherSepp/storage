package impl_test

import (
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"sync"
	"testing"
)

func TestNewCache(t *testing.T) {
	var buf []byte
	for _, size := range []int{-1, 0, 1, 50} {

		// new test cache (min. size == 17 MB)
		c := impl.NewCache(size)

		// test byte pool (use buf var for this)
		// get (min. 300)
		for i := 0; i < 1000; i++ {
			buf = c.Pool().Get()
			if buf == nil || len(buf) != interf.SectorSize {
				t.Fatalf("invalid buffer size")
			}
		}
		// set (min 300)
		for i := 0; i < 1000; i++ {
			c.Pool().Put(buf)
		}

		// check nil for next steps
		if buf == nil {
			t.Fatalf("invalid buffer")
		}

		// test Set()
		buf[0] = 0xff
		err := c.Set("fileId", 13, buf)
		if err != nil {
			t.Fatalf("%v", err)
		}

		// test Get()
		b, err := c.Get("fileId", 13, nil)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if b[0] != 0xff {
			t.Fatalf("invalid data")
		}

		// test data changes
		// init buf and set
		buf[0] = 10
		buf[1] = 11
		buf[2] = 12
		_ = c.Set("fileId", 99, buf)
		// mod buf
		buf[0] = 20
		buf[1] = 21
		buf[2] = 22
		// get buf with old data
		b, _ = c.Get("fileId", 99, c.Pool().Get())
		if b[0] != 10 || b[1] != 11 || b[2] != 12 {
			t.Fatalf("invalid data")
		}
		// mod buf from cache
		b[0] = 30
		b[1] = 31
		b[2] = 32
		// get buf (again) with old data
		b, _ = c.Get("fileId", 99, c.Pool().Get())
		if b[0] != 10 || b[1] != 11 || b[2] != 12 {
			t.Fatalf("invalid data")
		}
	}
}

//--------------------------------------------------------------------------------------------------------------------//

func TestRace_Cache(t *testing.T) {
	c := impl.NewCache(0)

	var wg sync.WaitGroup
	wg.Add(5)
	for n := 0; n < 5; n++ {
		go func() {
			//------------------------------
			for i := 0; i < 1000; i++ {
				errS := c.Set("fileId", uint64(i), []byte{0xff})
				b, errG := c.Get("fileId", uint64(i), nil)
				if errS != nil || errG != nil || len(b) != 1 || b[0] != 0xff {
					t.Fail()
				}
			}
			//------------------------------
			wg.Done()
		}()
	}
	wg.Wait()
}
