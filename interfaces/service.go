package interf

import (
	"io"
)

// Service is the central interface to access the storage.
type Service interface {

	// Update the internal file index, which can be accessed with Files().
	// Only files in the configured root directory (see parentFolderId) are processed.
	// Folders and sub-folders are ignored. This method is very slow at the first call!
	// This method is thread-safe.
	Update() error

	// Files returns all available files.
	// This method is offline and does not trigger a connection to the storage.
	// The internal file index must be updated separately with Update().
	// The map key is the file id.
	// This method is thread-safe.
	Files() Files

	// Save reads bytes from the io.Reader r and saves them in the storage.
	// The file name can exist multiple times and existing files with the same name are not overwritten.
	// The param max limits the read bytes (see io.LimitedReader). max=0 means read until EOF.
	// Return the new file id of the saved file (if successful).
	// Don't forget to call Update().
	// This method is thread-safe.
	Save(name string, r io.Reader, max int64) (file File, err error)

	// Trash moves a file (identified via the file id) to the trash.
	// Don't forget to call Update().
	// This method is thread-safe.
	Trash(file File) error

	// Reader enables read access to a file identified by the file id.
	// The connection must be closed manually with Close() after use.
	// This method is thread-safe.
	Reader(file File, off int64) (io.ReadCloser, error)

	// LimitedReader enables read access to a file identified by the file id,
	// but stops with EOF after n bytes. This method behaves like io.LimitedReader.
	// The connection must be closed manually with Close() after use.
	// This method is thread-safe.
	LimitedReader(file File, off int64, n int64) (io.ReadCloser, error)

	// ReaderAt allow random read access to a file identified by the file id.
	// A cache must be used internally for random read access.
	// It may also be necessary to open several internal connections to the storage.
	// The connection must be closed manually with Close() after use.
	// This method is thread-safe.
	ReaderAt(file File) (ReaderAt, error)

	// MultiReaderAt allow random read access to a series of files identified by the file ids.
	// All files except the last file must be the same size.
	// In addition, this method behaves like ReaderAt.
	//
	// A cache must be used internally for random read access.
	// It may also be necessary to open several internal connections to the storage.
	// The connection must be closed manually with Close() after use.
	// This method is thread-safe.
	MultiReaderAt(list []File) (ReaderAt, error)
}
