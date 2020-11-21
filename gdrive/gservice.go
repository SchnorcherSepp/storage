package gdrive

import (
	"bytes"
	"errors"
	"fmt"
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	google "google.golang.org/api/drive/v3"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"
)

// interface check: interf.Service
var _ interf.Service = (*_GService)(nil)

// _GService the central interface to access the Google Drive storage.
// Must be created with NewService().
type _GService struct {
	google         *google.Service
	parent         string
	cacheFile      string
	readerCache    interf.Cache
	debugLvl       uint8
	mux            *sync.RWMutex
	files          interf.Files
	initialized    bool
	startPageToken string
	skipFullInit   bool
}

// NewGService returns an interface to Google Drive. The parent specifies the folder
// with the active files. If the value is "root" or empty, the root directory of Google Drive is used.
// indexCacheFile is used to speed up Update (files are available faster).
// With skipFullInit = true, the init update call ends with a successful loading of indexCacheFile.
// readerCache=nil disable the cache for ReaderAt() and MultiReaderAt()
// debugLvl (@see impl.DebugHigh and impl.DebugOff)
func NewGService(parent, indexCacheFile string, skipFullInit bool, oauth *google.Service, readerCache interf.Cache, debugLvl uint8) interf.Service {
	s := &_GService{
		google:         oauth,
		parent:         parent,
		cacheFile:      indexCacheFile,
		readerCache:    readerCache,
		debugLvl:       debugLvl,
		mux:            new(sync.RWMutex),
		files:          impl.NewFiles(nil), // empty list, set by Update()
		initialized:    false,
		startPageToken: "",
		skipFullInit:   skipFullInit,
	}

	// root fix: replace root alias with valid folder id
	if s.parent == "root" || s.parent == "" {
		root, err := s.google.Files.Get("root").Do()
		if err != nil {
			// do nothing
			log.Printf("ERROR: %s/rootFix: %v", packageName, err)
		} else {
			// update parent folder id
			log.Printf("INFO: %s/rootFix: change parent folder id '%s' to '%s'", packageName, s.parent, root.Id)
			s.parent = root.Id
		}
	}
	return s
}

//--------------------------------------------------------------------------------------------------------------------//

// Update is the implementation of Service.Update()
//
// Update the internal file index, which can be accessed with Files().
// Only files in the configured root directory (see parentFolderId) are processed.
// Folders and sub-folders are ignored. This method is very slow on the first call!
// This method is thread-safe.
func (s *_GService) Update() error {
	s.mux.RLock() // READ Lock
	bl := s.initialized
	s.mux.RUnlock()

	if bl {
		return s.updateFiles() // thread safe
	} else {
		return s.initFiles() // thread safe
	}
}

// Files is the implementation of Service.Files()
//
// Files returns a cloned map of all available files.
// This method is offline and does not trigger a connection to (google) drive.
// The internal file index must be updated separately with Update().
// The map key is the file id.
// This method is thread-safe.
func (s *_GService) Files() interf.Files {
	s.mux.RLock() // READ Lock
	defer s.mux.RUnlock()

	return s.files
}

// Save is the implementation of Service.Save()
// The Google API does not set the values 'size' and 'md5'!
//
// Save reads bytes from the io.Reader r and saves them in (google) drive.
// The file name can exist multiple times and existing files with the same name are not overwritten.
// max limits the number of bytes to be read (see io.LimitedReader). max=0 means read until EOF.
// Return the new file id of the saved file (if successful).
// Don't forget to call Update().
// This method is thread-safe.
func (s *_GService) Save(name string, r io.Reader, max int64) (file interf.File, err error) {
	name = strings.TrimSpace(name)
	if name == "" || r == nil {
		return nil, errors.New("invalid input")
	}

	// set google file metadata
	f := &google.File{
		Name:     name,
		Parents:  []string{s.parent},
		MimeType: "application/octet-stream",
	}

	// upload
	if max > 0 {
		r = io.LimitReader(r, max)
	}
	f, err = s.google.Files.Create(f).Media(r).Do()

	// request error
	if err != nil {
		errMsg := fmt.Sprintf("%v", err)
		if strings.Contains(errMsg, "insufficientPermissions") {
			// wrong permissions
			return nil, fmt.Errorf("upload error: wrong permissions: create a new oauth token with write permissions: %v", err)
		} else {
			// other error
			return nil, fmt.Errorf("upload error: %v", err)
		}
	}

	// success
	// The Google API does not set the values 'size' and 'md5'!
	return impl.NewFile(f.Id, f.Name, ParseTime(time.Now().Format("2006-01-02T15:04:05.700Z")), f.Size, f.Md5Checksum), nil
}

// Trash is the implementation of Service.Trash()
//
// Trash moves a file (identified via the file id) to the trash.
// Don't forget to call Update().
// This method is thread-safe.
func (s *_GService) Trash(file interf.File) error {
	id := ""
	if file != nil {
		id = file.Id()
	}

	_, err := s.google.Files.Update(id, &google.File{Trashed: true}).Do() // thread safe
	return err
}

// Reader is the implementation of Service.Reader()
// Delegate to LimitedReader with n=interf.MaxFileSize
func (s *_GService) Reader(file interf.File, off int64) (io.ReadCloser, error) {
	return s.LimitedReader(file, off, interf.MaxFileSize)
}

// LimitedReader is the implementation of Service.LimitedReader()
//
// Reader enables read access to a file identified by the file id.
// The connection must be closed manually with Close() after use.
// This method is thread-safe.
func (s *_GService) LimitedReader(file interf.File, off int64, n int64) (io.ReadCloser, error) {
	if n < 1 {
		// n = 0 -> no data requested -> return nothing
		return ioutil.NopCloser(bytes.NewReader([]byte{})), nil
	}

	id := ""
	if file != nil {
		id = file.Id()
	}

	get := s.google.Files.Get(id)
	get.Header().Set("Range", fmt.Sprintf("bytes=%d-%d", off, off+n-1))

	resp, err := get.Download()
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ReaderAt is the implementation of Service.ReaderAt()
func (s *_GService) ReaderAt(file interf.File) (interf.ReaderAt, error) {
	return impl.NewReaderAt(file, s, s.readerCache, s.debugLvl)
}

// MultiReaderAt is the implementation of Service.MultiReaderAt()
func (s *_GService) MultiReaderAt(list []interf.File) (interf.ReaderAt, error) {
	if len(list) == 1 {
		// use the normal ReaderAt for single files
		return s.ReaderAt(list[0])
	} else {
		// MultiReaderAt
		return impl.NewMultiReaderAt(list, s, s.readerCache, s.debugLvl)
	}
}

// Cache returns the internal cache instance. Can be NIL.
func (s *_GService) Cache() interf.Cache {
	return s.readerCache
}

//---------  Helper  -------------------------------------------------------------------------------------------------//

// initFiles updates the internal indexcache with all FILES from the defined folder (parent folder id).
// Folders and files from sub folders are ignored. This method can be VERY SLOW, but must be called
// at least once when the program starts! After that you should work with updateFiles().
// To speed up the initialization at program start, data from the indexcache file can be used.
func (s *_GService) initFiles() error {

	// use indexcache
	// loading the last state allows to speed up the process
	s.mux.Lock() // <-------------- LOCK
	err := cacheLoad(s)
	s.mux.Unlock() // <------------ UNLOCK

	if err != nil {
		log.Printf("WARNING: %s/initFiles: cacheLoad() failed: %v", packageName, err)
	}

	// try updateFiles() to validate indexcache files and indexcache startPageToken and get all updates
	s.mux.RLock() // <-------------- R LOCK
	token := s.startPageToken
	s.mux.RUnlock() // <------------ R UNLOCK

	if token != "" {
		err := s.updateFiles() // thread safe
		if err != nil {
			log.Printf("ERROR: %s/initFiles: UpdateFileList() failed with indexcache: %v", packageName, err)
		} else {
			log.Printf("INFO: %s/initFiles: speed up initialization with indexcache", packageName)
			// fast init (optional)
			if s.skipFullInit {
				s.mux.Lock() // <-------------- LOCK
				s.initialized = true
				s.mux.Unlock() // <------------ UNLOCK
				log.Printf("INFO: %s/initFiles: skip full initialisation", packageName)
				return nil // EXIT
			}
		}
	} else {
		log.Printf("INFO: %s/initFiles: initialization without indexcache", packageName)
	}

	// --- The function rebuilds the list no matter what. But valid Files are nice to have at this point --- //

	// config
	const folderMimeType = "application/vnd.google-apps.folder"
	const fields = "nextPageToken, files(id, name, size, modifiedTime, md5Checksum)"
	const spaces = "drive" // Supported values are 'drive', 'appDataFolder' and 'photos'.
	const corpora = "user" // The user corpus includes all files in "My Drive" and "Shared with me"
	const pageSize = 1000  // split big file lists in pages (default 1000)
	query := fmt.Sprintf("trashed = false and mimeType != '%s' and '%s' in parents", folderMimeType, s.parent)

	// get a new StartPageToken to watch changes
	startPageTokenObj, err := s.google.Changes.GetStartPageToken().Do() // thread safe
	if err != nil {
		log.Printf("ERROR: %s/initFiles: can't get StartPageToken: %v", packageName, err)
		return err
	}
	s.mux.Lock() // <-------------- LOCK
	s.startPageToken = startPageTokenObj.StartPageToken
	s.mux.Unlock() // <------------ UNLOCK

	// get all relevant files
	newList := make(map[string]interf.File)
	pageToken := ""
	for {
		// read a result page
		fileList, err := s.google.Files.List().Q(query).PageToken(pageToken).
			Spaces(spaces).Corpora(corpora).PageSize(int64(pageSize)).
			Fields(fields).Do() // thread safe

		// error handling
		if err != nil {
			log.Printf("ERROR: %s/initFiles: can't read all result pages: %v", packageName, err)
			return err
		}

		// add all results (files) to list
		for _, f := range fileList.Files {
			newList[f.Id] = impl.NewFile(f.Id, f.Name, ParseTime(f.ModifiedTime), f.Size, f.Md5Checksum)
		}

		// break loop (no more pages)
		pageToken = fileList.NextPageToken
		if pageToken == "" {
			log.Printf("INFO: %s/initFiles: successful files initialization (%d files)", packageName, len(newList))
			break
		}
	}

	// FIN: set new list, save indexcache and return
	s.mux.Lock() // <-------------- LOCK
	s.files = impl.NewFiles(newList)
	s.initialized = true
	err = cacheSave(s, s.files)
	s.mux.Unlock() // <------------ UNLOCK

	if err != nil {
		log.Printf("ERROR: %s/initFiles: cacheSave() failed: %v", packageName, err)
	}
	return nil
}

//--------------------------------------------------------------------------------------------------------------------//

// updateFiles only queries a delta of the internal file list of google drive.
// This makes the function much faster than initFiles().
func (s *_GService) updateFiles() error {

	// check startPageToken
	s.mux.RLock() // <-------------- R LOCK
	pageToken := s.startPageToken
	s.mux.RUnlock() // <------------ R UNLOCK

	if pageToken == "" {
		// invalid startPageToken
		s.mux.Lock() // <-------------- LOCK
		s.initialized = false
		s.mux.Unlock() // <------------ UNLOCK
		// return error
		msg := "can't use fast updateFiles() without StartPageToken, please call initFiles()"
		log.Printf("ERROR: %s/updateFiles: %s", packageName, msg)
		return fmt.Errorf(msg)
	}

	// config
	const folderMimeType = "application/vnd.google-apps.folder"
	const fields = "nextPageToken, newStartPageToken, changes(file(id, name, size, trashed, mimeType, parents, modifiedTime, md5Checksum))"
	const spaces = "drive" // Supported values are 'drive', 'appDataFolder' and 'photos'.
	const pageSize = 1000  // split big file lists in pages (default 1000)

	fileList := make(map[string]interf.File)
	for _, v := range s.Files().All() { // thread safe
		fileList[v.Id()] = v
	}

	// loop to get all changes
	for {
		// read a result pages
		changeList, err := s.google.Changes.List(pageToken).Spaces(spaces).PageSize(int64(pageSize)).Fields(fields).Do() // thread safe
		if err != nil {
			log.Printf("ERROR: %s/updateFiles: can't read all result pages: %v", packageName, err)
			return err
		}

		// update fileList
		for _, change := range changeList.Changes {
			// object on changeList is a file
			if change.File.MimeType != folderMimeType {
				// file is in the watched folder
				for _, fileParent := range change.File.Parents {
					if fileParent == s.parent { // s.parent is thread safe (no write access)
						// add/update or remove?
						if change.File.Trashed {
							// the change is: remove
							delete(fileList, change.File.Id)
						} else {
							// the change is: update or new file
							cf := change.File
							fileList[cf.Id] = impl.NewFile(cf.Id, cf.Name, ParseTime(cf.ModifiedTime), cf.Size, cf.Md5Checksum)
						}
					}
				}
			}
		}

		// break loop (no more pages)
		pageToken = changeList.NextPageToken // NextPageToken for the next page
		if pageToken == "" {
			// no more pages
			// set the new NewStartPageToken for the next updateFiles() call
			pageToken = changeList.NewStartPageToken
			log.Printf("INFO: %s/updateFiles: successful file update (%d files)", packageName, len(fileList))
			break
		}
	}

	//-----  THREAD SAFE  ----------------------------------------------------------------------
	s.mux.Lock() // LOCK
	defer s.mux.Unlock()

	s.startPageToken = pageToken
	s.files = impl.NewFiles(fileList)

	// write new state to indexcache file
	err := cacheSave(s, s.files)
	if err != nil {
		log.Printf("ERROR: %s/updateFiles: cacheSave() failed: %v", packageName, err)
	}
	return nil
}
