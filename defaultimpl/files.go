package impl

import (
	interf "github.com/SchnorcherSepp/storage/interfaces"
)

// interface check: interf.Files
var _ interf.Files = (*_Files)(nil)

// @see interf.Files
//
// Files maintains an internal list of files.
// Files and File are immutable objects!
type _Files struct {
	byId map[string]interf.File // this map is never nil (see NewFiles)
	list []interf.File          // set by NewFiles
}

// NewFiles return the default implementation of interf.Files.
// If the map is nil, a valid empty map is generated.
func NewFiles(byId map[string]interf.File) interf.Files {
	// convert nil to empty map
	if byId == nil {
		byId = make(map[string]interf.File)
	}

	// build list
	list := make([]interf.File, 0, len(byId))
	for _, f := range byId {
		if f != nil { // ignore nil elements
			list = append(list, f)
		}
	}

	// return
	return &_Files{
		byId: byId,
		list: list,
	}
}

// @see interf.Files
//
// All returns a list of all internally managed files.
// The list is created with every call and can be changed safely.
// There are no online connections to the storage (internal data are used).
// This method is thread safe (Files is an immutable object).
func (fs _Files) All() []interf.File {
	// return clone, not the inner list!
	list := make([]interf.File, len(fs.list))
	copy(list, fs.list) // clone list
	return list
}

// @see interf.Files
//
// ById returns the file with the requested file id.
// If no file is found, the os.ErrNotExist error is returned.
// There are no online connections to the storage (internal data are used).
// This method is thread safe (Files is an immutable object).
func (fs _Files) ById(id string) (interf.File, error) {
	return FileById(fs.byId, id) // redirect to FileById
}

// @see interf.Files
//
// ByName returns the latest (File.ModTime) file found with the requested name.
// If no file is found, the os.ErrNotExist error is returned.
// There are no online connections to the storage (internal data are used).
// This method is thread safe (Files is an immutable object).
func (fs _Files) ByName(name string) (interf.File, error) {
	return FileByName(fs.list, name) // redirect to FileByName
}

// @see interf.Files
//
// ByAttr returns the first file found with the requested attributes.
// If the parameter md5 is empty, this attribute is not considered in the search.
// If no file is found, the os.ErrNotExist error is returned.
// There are no online connections to the storage (internal data are used).
// This method is thread safe (Files is an immutable object).
func (fs _Files) ByAttr(name string, size int64, md5 string) (interf.File, error) {
	return FileByAttr(fs.list, name, size, md5) // redirect to FileByAttr
}
