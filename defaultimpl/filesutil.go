package impl

import (
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"math"
	"os"
)

// FileById returns the file with the requested file id.
// If no file is found, the os.ErrNotExist error is returned.
// The data source is the specified list of files (attribute files).
func FileById(files map[string]interf.File, id string) (interf.File, error) {
	// a NULL list cannot contain a file
	if files == nil {
		return nil, os.ErrNotExist
	}

	// get & return
	f, ok := files[id]
	if ok && f != nil {
		return f, nil // valid file found
	} else {
		return nil, os.ErrNotExist // nothing found or nil
	}
}

// FileByName returns the latest (File.ModTime) file found with the requested name.
// If no file is found, the os.ErrNotExist error is returned.
// The data source is the specified list of files (attribute files).
func FileByName(files []interf.File, name string) (interf.File, error) {
	// a NULL list cannot contain a file
	if files == nil {
		return nil, os.ErrNotExist
	}

	// find the latest file
	var ret interf.File = nil
	var age int64 = math.MinInt64
	var err = os.ErrNotExist

	for _, f := range files {
		if f != nil && f.Name() == name && f.ModTime() > age {
			ret = f
			age = f.ModTime()
			err = nil
		}
	}

	// return
	return ret, err
}

// FileByAttr returns the first file found with the requested attributes.
// If the parameter md5 is empty, this attribute is not considered in the search.
// If no file is found, the os.ErrNotExist error is returned.
// The data source is the specified list of files (attribute files).
func FileByAttr(files []interf.File, name string, size int64, md5 string) (interf.File, error) {
	// a NULL list cannot contain a file
	if files == nil {
		return nil, os.ErrNotExist
	}

	// return the first matching file (md5 is optional)
	for _, f := range files {
		if f != nil && f.Name() == name && f.Size() == size && (md5 == "" || f.Md5() == md5) {
			return f, nil
		}
	}

	// no results
	return nil, os.ErrNotExist
}
