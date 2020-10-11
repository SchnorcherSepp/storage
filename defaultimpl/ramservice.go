package impl

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"
)

// interface check: interf.Service
var _ interf.Service = (*_RamService)(nil)

// @see interf.Service
//
// Service is the central interface to access the storage.
type _RamService struct {
	cache    interf.Cache
	debugLvl uint8
	hidden   interf.Files
	files    interf.Files
	data     map[string][]byte
	mux      *sync.RWMutex
}

// NewRamService return the RAM implementation of interf.Service.
// The data are only in RAM. This implementation is mainly for testing.
func NewRamService(cache interf.Cache, debugLvl uint8) interf.Service {
	return &_RamService{
		cache:    cache,
		debugLvl: debugLvl,
		hidden:   NewFiles(nil),
		files:    NewFiles(nil),
		data:     make(map[string][]byte),
		mux:      new(sync.RWMutex),
	}
}

//-----------  IMPLEMENTATION:  @see interf.Service  -----------------------------------------------------------------//

func (s *_RamService) Update() error {
	s.mux.Lock() // WRITE Lock
	defer s.mux.Unlock()

	s.files = s.hidden
	return nil
}

func (s *_RamService) Files() interf.Files {
	s.mux.RLock() // READ Lock
	defer s.mux.RUnlock()

	return s.files
}

func (s *_RamService) Save(name string, r io.Reader, max int64) (file interf.File, err error) {
	// check input
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, errors.New("empty name")
	}
	if r == nil {
		return nil, errors.New("nil reader")
	}

	// limit reader
	if max > 0 {
		r = io.LimitReader(r, max)
	}

	// read all bytes
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// calc md5
	hash := md5.New()
	hash.Write(data)
	h := fmt.Sprintf("%x", hash.Sum(nil))

	// create new file
	f := NewFile(genId(), name, time.Now().Unix(), int64(len(data)), h)

	// ----- update lists -----------------------------------

	s.mux.Lock() // WRITE Lock
	defer s.mux.Unlock()

	// list to map
	byId := make(map[string]interf.File)
	for _, v := range s.hidden.All() {
		byId[v.Id()] = v
	}

	// update
	byId[f.Id()] = f
	s.data[f.Id()] = data
	s.hidden = NewFiles(byId)

	return f, nil
}

func (s *_RamService) Trash(file interf.File) error {
	s.mux.Lock() // WRITE Lock
	defer s.mux.Unlock()

	// list to map
	byId := make(map[string]interf.File)
	for _, v := range s.hidden.All() {
		byId[v.Id()] = v
	}

	// update
	_, ok := byId[file.Id()]
	if !ok {
		return errors.New("id not found")
	}

	delete(byId, file.Id())
	s.hidden = NewFiles(byId)

	return nil
}

func (s *_RamService) Reader(file interf.File, off int64) (io.ReadCloser, error) {
	s.mux.RLock() // READ Lock
	defer s.mux.RUnlock()

	// get file if exist
	f, err := s.hidden.ById(file.Id())
	if err != nil {
		return nil, err
	}

	// check offset
	data := s.data[f.Id()]
	if off >= int64(len(data)) {
		return nil, io.EOF
	}

	// return
	return ioutil.NopCloser(bytes.NewReader(data[off:])), nil
}

func (s *_RamService) LimitedReader(file interf.File, off int64, n int64) (io.ReadCloser, error) {
	r, err := s.Reader(file, off)

	if n > 0 && r != nil {
		r = ioutil.NopCloser(io.LimitReader(r, n))
	}

	return r, err
}

func (s *_RamService) ReaderAt(file interf.File) (interf.ReaderAt, error) {
	return NewReaderAt(file, s, s.cache, s.debugLvl)
}

func (s *_RamService) MultiReaderAt(list []interf.File) (interf.ReaderAt, error) {
	if len(list) == 1 {
		// use the normal ReaderAt for single files
		return s.ReaderAt(list[0])
	} else {
		// MultiReaderAt
		return NewMultiReaderAt(list, s, s.cache, s.debugLvl)
	}
}

func (s *_RamService) Cache() interf.Cache {
	return s.cache
}

//--------  Helper  --------------------------------------------------------------------------------------------------//

// genId generate a random (unique) fileId for new files.
func genId() string {
	buf := make([]byte, 32)

	n, err := rand.Read(buf)
	if err != nil {
		log.Printf("ERROR: impl/genId: read error: %v", err)
		return "1er-ijL8xxErRoRxxN-ZKxxErRoRxWl9C"
	}

	s := base64.URLEncoding.EncodeToString(buf[:n])
	if len(s) < 30 {
		log.Printf("ERROR: impl/genId: len error: %d", len(s))
		return "1er-ojo8xxErRoRxxN-oKxxErRoRxWloC"
	}

	return "1pl-" + s[0:14] + "-" + s[14:28]
}
