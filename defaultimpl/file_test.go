package impl_test

import (
	"fmt"
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	"sync"
	"testing"
)

func TestNewFile(t *testing.T) {
	// test variables
	id := "fileId"
	name := "name.file"
	modTime := int64(1584535538)
	size := int64(16317)
	md5 := "098f6bcd4621d373c0de4e832627b4f6"

	// test NewFile()
	f := impl.NewFile(id, name, modTime, size, md5)
	if f == nil {
		t.Fatalf("NewFile returns nil")
	}

	// test getter
	if f.Id() != id {
		t.Errorf("%s != %s", f.Id(), id)
	}
	if f.Name() != name {
		t.Errorf("%s != %s", f.Name(), name)
	}
	if f.ModTime() != modTime {
		t.Errorf("%d != %d", f.ModTime(), modTime)
	}
	if f.Size() != size {
		t.Errorf("%d != %d", f.Size(), size)
	}
	if f.Md5() != md5 {
		t.Errorf("%s != %s", f.Md5(), md5)
	}
}

//--------------------------------------------------------------------------------------------------------------------//

func TestRace_File(t *testing.T) {
	f := impl.NewFile("fileId", "Name.file", 12345, 6789, "hex_hex_hex_hex")

	var wg sync.WaitGroup
	wg.Add(5)
	for n := 0; n < 5; n++ {
		go func() {
			//------------------------------
			for i := 0; i < 1000; i++ {
				s := fmt.Sprintf("%s, %s, %d, %d, %s", f.Id(), f.Name(), f.Size(), f.ModTime(), f.Md5())
				if s == "" {
					t.Fail()
				}
			}
			//------------------------------
			wg.Done()
		}()
	}
	wg.Wait()
}
