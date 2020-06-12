package interf

// Files maintains an internal list of files.
// Files and File are immutable objects!
type Files interface {

	// All returns a list of all internally managed files.
	// The list is created with every call and can be changed safely.
	// There are no online connections to the storage (internal data are used).
	// This method is thread safe (Files is an immutable object).
	All() []File

	// ById returns the file with the requested file id.
	// If no file is found, the os.ErrNotExist error is returned.
	// There are no online connections to the storage (internal data are used).
	// This method is thread safe (Files is an immutable object).
	ById(id string) (File, error)

	// ByName returns the latest (File.ModTime) file found with the requested name.
	// If no file is found, the os.ErrNotExist error is returned.
	// There are no online connections to the storage (internal data are used).
	// This method is thread safe (Files is an immutable object).
	ByName(name string) (File, error)

	// ByAttr returns the first file found with the requested attributes.
	// If the parameter md5 is empty, this attribute is not considered in the search.
	// If no file is found, the os.ErrNotExist error is returned.
	// There are no online connections to the storage (internal data are used).
	// This method is thread safe (Files is an immutable object).
	ByAttr(name string, size int64, md5 string) (File, error)
}
