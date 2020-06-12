package impl_test

import (
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"os"
	"testing"
)

func TestFileById(t *testing.T) {
	fileId := "fileId"
	file := impl.NewFile(fileId, "", 0, 0, "")

	// TEST: input (nil, map:nil, map:file)
	if f, err := impl.FileById(nil, fileId); f != nil || err != os.ErrNotExist {
		t.Fatalf("no error: f=%v, e=%v", f, err)
	}
	if f, err := impl.FileById(map[string]interf.File{"fileId": nil}, fileId); f != nil || err != os.ErrNotExist {
		t.Fatalf("no error: f=%v, e=%v", f, err)
	}
	if f, err := impl.FileById(map[string]interf.File{"fileId": file}, fileId); f != file || err != nil {
		t.Fatalf("no error: f=%v, e=%v", f, err)
	}
}

func TestFileByName(t *testing.T) {
	name := "name.file"
	file := impl.NewFile("", name, 0, 0, "")

	// TEST: input (nil, map:nil, map:file)
	if f, err := impl.FileByName(nil, name); f != nil || err != os.ErrNotExist {
		t.Fatalf("no error: f=%v, e=%v", f, err)
	}
	if f, err := impl.FileByName([]interf.File{nil}, name); f != nil || err != os.ErrNotExist {
		t.Fatalf("no error: f=%v, e=%v", f, err)
	}
	if f, err := impl.FileByName([]interf.File{file}, name); f != file || err != nil {
		t.Fatalf("no error: f=%v, e=%v", f, err)
	}
}

func TestFileByAttr(t *testing.T) {
	name := "name.file"
	size := int64(123)
	md5 := "hash123"
	file := impl.NewFile("", name, 0, size, md5)

	// TEST: input (nil, map:nil, map:file)
	if f, err := impl.FileByAttr(nil, name, size, md5); f != nil || err != os.ErrNotExist {
		t.Fatalf("no error: f=%v, e=%v", f, err)
	}
	if f, err := impl.FileByAttr([]interf.File{nil}, name, size, md5); f != nil || err != os.ErrNotExist {
		t.Fatalf("no error: f=%v, e=%v", f, err)
	}
	if f, err := impl.FileByAttr([]interf.File{file}, name, size, md5); f != file || err != nil {
		t.Fatalf("no error: f=%v, e=%v", f, err)
	}
}
