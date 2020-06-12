package interf

// File stands for a single file in storage.
// File is an immutable object!
type File interface {

	// Id uniquely identifies a file.
	// This method is thread safe (File is an immutable object).
	// Example: 1pl-ijL8cnNcS2mBwN-ZKxHYUdL3DTl9C
	Id() string

	// Name of the file. The name is not unique and there can be multiple files with the same name.
	// This method is thread safe (File is an immutable object).
	// Example: test.dat
	Name() string

	// ModTime show the last change or update of the object (unix time; seconds).
	// If a file has never been changed, it's the time of creation.
	// This method is thread safe (File is an immutable object).
	// Example: 1584535538
	ModTime() int64

	// Size is the file size in bytes.
	// This method is thread safe (File is an immutable object).
	// Example 16317
	Size() int64

	// Md5 is the hash of the file content (hex string).
	// This method is thread safe (File is an immutable object).
	// Example: 098f6bcd4621d373c0de4e832627b4f6
	Md5() string
}
