package impl_test

import (
	"fmt"
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"sync"
	"testing"
)

func TestNewFiles(t *testing.T) {
	// test file variables
	id := "fileId"
	name := "name.file"
	modTime := int64(1584535538)
	size := int64(16317)
	md5 := "098f6bcd4621d373c0de4e832627b4f6"

	// param is NULL or EMPTY
	for _, param := range []map[string]interf.File{nil, {}} {
		fs := impl.NewFiles(param)
		if fs == nil {
			t.Fatalf("NewFile returns nil")
		}
		if len(fs.All()) != 0 {
			t.Errorf("wrong All()")
		}
		if f, err := fs.ById(id); err == nil || f != nil {
			t.Errorf("wrong ById()")
		}
		if f, err := fs.ByName(name); err == nil || f != nil {
			t.Errorf("wrong ByName()")
		}
		if f, err := fs.ByAttr(name, size, md5); err == nil || f != nil {
			t.Errorf("wrong ByAttr()")
		}
	}

	// param is OK
	const invalidId = "invalid"
	fs := impl.NewFiles(map[string]interf.File{
		invalidId: nil,
		"tmp1":    impl.NewFile("tmp1", name, 123, size, "aaa"),
		id:        impl.NewFile(id, name, modTime, size, md5),
		"tmp2":    impl.NewFile("tmp2", name, 456, size, "bbb"),
	})
	if fs == nil {
		t.Fatalf("NewFile returns nil")
	}
	if len(fs.All()) != 3 { // 3 because nil elements are ignored
		t.Errorf("wrong All()")
	}
	if f, err := fs.ById(invalidId); err == nil || f != nil {
		t.Errorf("wrong ById()")
	}

	// test ById
	if f, err := fs.ById(id); err != nil || f == nil || f.Id() != id {
		t.Errorf("wrong ById()")
	}

	// test ByName
	if f, err := fs.ByName(name); err != nil || f == nil || f.Id() != id {
		t.Errorf("wrong ByName()")
	}

	// test ByAttr
	if f, err := fs.ByAttr(name, size, md5); err != nil || f == nil || f.Id() != id {
		t.Errorf("wrong ByAttr()")
	}

	// test ByAttr (no md5)
	if f, err := fs.ByAttr(name, size, "ERROR"); err == nil || f != nil {
		t.Errorf("wrong ByAttr(): %v", f)
	}
	if f, err := fs.ByAttr(name, size, ""); err != nil || f == nil {
		t.Errorf("wrong ByAttr(): %v", f)
	}
}

//--------------------------------------------------------------------------------------------------------------------//

func TestRace_Files(t *testing.T) {
	byId := make(map[string]interf.File)
	byId["test"] = nil
	fs := impl.NewFiles(byId)

	var wg sync.WaitGroup
	wg.Add(5)
	for n := 0; n < 5; n++ {
		go func() {
			//------------------------------
			for i := 0; i < 1000; i++ {
				a := fs.All()
				b, _ := fs.ById("")
				c, _ := fs.ByName("")
				d, _ := fs.ByAttr("", 0, "")
				s := fmt.Sprintf("%s, %s, %s, %s", a, b, c, d)
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
