package impl

import (
	interf "github.com/SchnorcherSepp/storage/interfaces"
)

// interface check: interf.File
var _ interf.File = (*_File)(nil)

// @see interf.File
//
// File stands for a single file in storage.
// File is an immutable object!
type _File struct {
	id      string
	name    string
	modTime int64
	size    int64
	md5     string
}

// NewFile return the default implementation of interf.File.
// This encapsulates the given data.
func NewFile(id, name string, modTime, size int64, md5 string) interf.File {
	return &_File{
		id:      id,
		name:    name,
		modTime: modTime,
		size:    size,
		md5:     md5,
	}
}

// @see interf.File
//
// Id uniquely identifies a file.
// This method is thread safe (File is an immutable object).
// Example: 1pl-ijL8cnNcS2mBwN-ZKxHYUdL3DTl9C
func (f *_File) Id() string {
	return f.id
}

// @see interf.File
//
// Name of the file. The name is not unique and there can be multiple files with the same name.
// This method is thread safe (File is an immutable object).
// Example: test.dat
func (f *_File) Name() string {
	return f.name
}

// @see interf.File
//
// ModTime show the last change or update of the object (unix time; seconds).
// If a file has never been changed, it's the time of creation.
// This method is thread safe (File is an immutable object).
// Example: 1584535538
func (f *_File) ModTime() int64 {
	return f.modTime
}

// @see interf.File
//
// Size is the file size in bytes.
// This method is thread safe (File is an immutable object).
// Example 16317
func (f *_File) Size() int64 {
	return f.size
}

// @see interf.File
//
// Md5 is the hash of the file content (hex string).
// This method is thread safe (File is an immutable object).
// Example: 098f6bcd4621d373c0de4e832627b4f6
func (f *_File) Md5() string {
	return f.md5
}
