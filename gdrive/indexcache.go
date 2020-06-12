package gdrive

import (
	"crypto/md5"
	"encoding/gob"
	"errors"
	"fmt"
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"os"
)

// IndexCache stores the last valid drive file list.
// Loading the last state allows you to view data while building a new list.
// You can update a valid file list state with the StartPageToken very fast (load diff)
type _IndexCache struct {
	Files          map[string]_File
	StartPageToken string
	CacheSig       string
}

// File is a helper with exported attributes for serialization.
type _File struct {
	Id      string
	Name    string
	ModTime int64
	Size    int64
	Md5     string
}

//--------------------------------------------------------------------------------------------------------------------//

// cacheSave save the file list and the StartPageToken to a file
func cacheSave(s *_GService, files interf.Files) error {

	// create file list for serialization
	list := make(map[string]_File)
	for _, f := range files.All() { // thread safe
		list[f.Id()] = _File{
			Id:      f.Id(),
			Name:    f.Name(),
			ModTime: f.ModTime(),
			Size:    f.Size(),
			Md5:     f.Md5(),
		}
	}

	// calc cache sig
	cacheSig, err := cacheSig(s) // thread safe (connection to google server)
	if err != nil {
		return err
	}

	// create IndexCache
	indexCache := _IndexCache{
		Files:          list,
		StartPageToken: s.startPageToken, // NOT thread safe !!!
		CacheSig:       cacheSig,
	}

	// create new indexcache file
	fh, err := os.OpenFile(s.cacheFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) // thread safe ??
	if err != nil {
		return err
	}
	defer fh.Close()

	// write data (serialize indexcache)
	if err := gob.NewEncoder(fh).Encode(indexCache); err != nil {
		return err
	}

	// success
	return nil
}

// cacheLoad load the last valid drive Files from indexcache file and set file list and startPageToken.
func cacheLoad(s *_GService) error {

	// exists indexcache file?
	if _, err := os.Stat(s.cacheFile); err != nil {
		return err
	}

	// open indexcache file
	fh, err := os.Open(s.cacheFile)
	if err != nil {
		return err
	}
	defer fh.Close()

	// load indexcache object
	indexCache := new(_IndexCache)
	if err := gob.NewDecoder(fh).Decode(&indexCache); err != nil {
		return err
	}

	// calc cache sig
	cacheSig, err := cacheSig(s)
	if err != nil {
		return err
	}

	// check indexcache signature
	if indexCache.CacheSig != cacheSig {
		return errors.New("wrong indexcache signature")
	}

	// create interf.Files
	byId := make(map[string]interf.File)
	for k, v := range indexCache.Files {
		byId[k] = impl.NewFile(v.Id, v.Name, v.ModTime, v.Size, v.Md5)
	}
	files := impl.NewFiles(byId)

	// set indexcache data
	s.files = files
	s.startPageToken = indexCache.StartPageToken

	return nil
}

//--------  HELPER  --------------------------------------------------------------------------------------------------//

// cacheSig bind the indexcache to the oauth file and the rootFolderId.
// Any change to the cacheSig invalidates the indexcache.
// This function needs an active connection to the google server.
func cacheSig(s *_GService) (string, error) {
	const errorSig = "ERROR-SIG"

	// get PermissionId: The user's ID as visible in Permission resources (oauth file).
	about, err := s.google.About.Get().Fields("user(permissionId)").Do()
	if err != nil {
		return errorSig, err
	}
	permId := about.User.PermissionId

	// check PermissionId
	if len(permId) < 3 {
		return errorSig, errors.New("invalid user permissionId")
	}

	// calc sig
	h := md5.New()
	h.Write([]byte(s.parent)) // parent folder (example 'root')
	h.Write([]byte("|"))
	h.Write([]byte(permId)) // permId (= google user)

	return fmt.Sprintf("%x", h.Sum(nil)), nil // return cacheSig
}
