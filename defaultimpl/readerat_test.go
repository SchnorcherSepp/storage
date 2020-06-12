package impl_test

import (
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"io"
	"log"
	"sync"
	"testing"
)

func TestNewReaderAt(t *testing.T) {
	f, s, _ := initTestFileAndTestService(t)

	// test with invalid file and invalid service
	if _, err := impl.NewReaderAt(nil, s, nil, true); err == nil {
		t.Fatal("no error with invalid file")
	}
	if _, err := impl.NewReaderAt(f, nil, nil, true); err == nil {
		t.Fatal("no error with invalid file")
	}

	// test without cache
	_, err := impl.NewReaderAt(f, s, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	// test with cache
	c := impl.NewCache(1)
	_, err = impl.NewReaderAt(f, s, c, true)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ReaderAt_ReadAt__without_cache(t *testing.T) {
	f, s, _ := initTestFileAndTestService(t)

	// ----------------- test without cache (for more internal tests) ---------------------------
	r, err := impl.NewReaderAt(f, s, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	ts := &testStat{t: t, at: r}

	// test READ: empty or invalid buffer (= zero data request) ---------------------------------
	if n, err := r.ReadAt(nil, -1); n != 0 || err != nil {
		t.Fatalf("ERROR: %v (n=%d)", err, n)
	}
	if n, err := r.ReadAt(make([]byte, 0), -1); n != 0 || err != nil {
		t.Fatalf("ERROR: %v (n=%d)", err, n)
	}

	// CHECK internal activities
	ts.RAtNew++ // NewReaderAt() is called    !!!  ReadAt with invalid buffer don't count !!!
	ts.Check()  //--------------------------------------------------------------------------------

	// test READ: request 1 byte AND invalid offset
	b := make([]byte, 1)
	if n, err := r.ReadAt(b, -1); n != 1 || err != nil || b[0] != 38 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++       // one request: ReadAt()
	ts.RAtAdd++       // no open reader (add one new)
	ts.RAtSectorRet++ // req in one new sector
	ts.Check()        //--------------------------------------------------------------------------------

	// test READ: request next byte (same sector; no cache!)
	if n, err := r.ReadAt(b, 1); n != 1 || err != nil || b[0] != 197 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++       // request: ReadAt()
	ts.RAtSectorRet++ // we have no cache -> we have to read the sector again
	ts.RAtAdd++       // and the open reader can't read the old sector again
	ts.Check()        //--------------------------------------------------------------------------------

	// test READ: request next SECTOR (use open reader)
	if n, err := r.ReadAt(b, interf.SectorSize); n != 1 || err != nil || b[0] != 108 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++       // request: ReadAt()
	ts.RAtBest++      // reuse open reader for next sector
	ts.RAtSectorRet++ // read next sector
	ts.Check()        //--------------------------------------------------------------------------------

	// test READ: skip sector 2 and sector 3 and read sector 4  (reuse open reader[s=1])
	if n, err := r.ReadAt(b, 4*interf.SectorSize); n != 1 || err != nil || b[0] != 71 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++        // request: ReadAt()
	ts.RAtBest++       // reuse open reader for sector
	ts.RAtSectorSkip++ // skip sector 3
	ts.RAtSectorSkip++ // skip sector 4
	ts.RAtSectorRet++  // read next sector
	ts.Check()         //--------------------------------------------------------------------------------

	// test READ: read bytes from two sectors
	b = make([]byte, interf.SectorSize)
	if n, err := r.ReadAt(b, interf.SectorSize/2); n != interf.SectorSize || err != nil || b[0] != 106 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++       // request: ReadAt()
	ts.RAtSectorRet++ // we have no cache -> we have to read the sector again
	ts.RAtAdd++       // and the open reader can't read the old sector again
	ts.RAtSectorRet++ // read two sectors
	ts.RAtBest++      // reuse open reader for next sector
	ts.Check()        //--------------------------------------------------------------------------------

	// test READ: jump to the last sector and read the last byte
	b = make([]byte, 1)
	if n, err := r.ReadAt(b, 150*1024*1024); n != 1 || err != io.EOF || b[0] != 254 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++       // request: ReadAt()
	ts.RAtAdd++       // we can't jump (too far away)
	ts.RAtSectorRet++ // read last sector
	ts.Check()        //--------------------------------------------------------------------------------

	// test READ: read over EOF
	b = make([]byte, 1)
	if n, err := r.ReadAt(b, 150*1024*1024+1); n != 0 || err != io.EOF || b[0] != 0 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++ // request: ReadAt()
	ts.RAtAdd++
	ts.RAtSectorRet++ // read last sector
	ts.Check()        //--------------------------------------------------------------------------------

	// test READ: read over EOF (special)
	// When ReadAt returns n < len(p), it returns a non-nil error
	// explaining why more bytes were not returned. In this respect,
	// ReadAt is stricter than Read.
	b = make([]byte, 3)
	if n, err := r.ReadAt(b, 150*1024*1024-1); n != 2 || err != io.EOF || b[0] != 39 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++       // request: ReadAt()
	ts.RAtAdd++       // no valid reader
	ts.RAtSectorRet++ // read last sector
	ts.RAtBest++      // reuse open reader for next sector
	ts.RAtSectorRet++ // return next sector (but this sector does not exist and return len=0 and EOF)
	ts.Check()        //--------------------------------------------------------------------------------

	// test READ: read in nowhere
	b = make([]byte, 33)
	if n, err := r.ReadAt(b, 150*1024*1024+77); n != 0 || err != io.EOF || b[0] != 0 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++
	ts.RAtAdd++
	ts.RAtSectorRet++
	ts.Check() //--------------------------------------------------------------------------------

	// PRINT STATS
	log.Printf("%#v", r.Stat())
}

func Test_ReaderAt_ReadAt__with_cache(t *testing.T) {
	f, s, _ := initTestFileAndTestService(t)
	c := impl.NewCache(1)

	r, err := impl.NewReaderAt(f, s, c, true)
	if err != nil {
		t.Fatal(err)
	}
	ts := &testStat{t: t, at: r}

	// READ: request second byte (sector 0) -----------------------------------------------------
	b := make([]byte, 1)
	if n, err := r.ReadAt(b, 1); n != 1 || err != nil || b[0] != 197 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtNew++       // init
	ts.RAtReq++       // one request: ReadAt()
	ts.CacheMis++     // ask cache first
	ts.RAtAdd++       // no open reader (add one new)
	ts.RAtSectorRet++ // req in one new sector
	ts.CacheSet++     // save sector
	ts.Check()        //--------------------------------------------------------------------------------

	// test READ: request first byte (same sector; = read back)
	if n, err := r.ReadAt(b, 1); n != 1 || err != nil || b[0] != 197 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++   // request: ReadAt()
	ts.CacheHit++ // use sector from cache
	ts.Check()    //--------------------------------------------------------------------------------

	// test READ: jump (and save skip-sectors)
	if n, err := r.ReadAt(b, 3*interf.SectorSize); n != 1 || err != nil || b[0] != 26 {
		t.Fatalf("ERROR: %v (n=%d, b=%v)", err, n, b)
	}

	// CHECK internal activities
	ts.RAtReq++        // request: ReadAt()
	ts.CacheMis++      // new sector
	ts.RAtBest++       // reuse open reader
	ts.RAtSectorSkip++ // skip sector 1
	ts.CacheSet++      // but save the sector in the cache
	ts.RAtSectorSkip++ // skip sector 2
	ts.CacheSet++      // but save the sector in the cache
	ts.RAtSectorRet++  // read sector 3
	ts.CacheSet++      // and save the sector in the cache
	ts.Check()         //--------------------------------------------------------------------------------

	// PRINT STATS
	log.Printf("%#v", r.Stat())
}

//--------------------------------------------------------------------------------------------------------------------//

func TestRace_ReaderAt(t *testing.T) {
	f, s, _ := initTestFileAndTestService(t)

	r, err := impl.NewReaderAt(f, s, nil, false) // test without cache for more inner code tests
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(5)
	for n := 0; n < 5; n++ {
		go func() {
			//------------------------------
			for i := 0; i < 1000; i++ {
				n, err1 := r.ReadAt(make([]byte, 1), int64(i))
				err2 := r.Close()
				r.Stat()
				if err1 != nil || err2 != nil || n != 1 {
					t.Fail()
				}
			}
			//------------------------------
			wg.Done()
		}()
	}
	wg.Wait()
}

//--------  HELPER  --------------------------------------------------------------------------------------------------//

func initTestFileAndTestService(t *testing.T) (interf.File, interf.Service, []interf.File) {
	s := impl.NewRamService(nil, false)
	if err := impl.InitDemo(s); err != nil {
		t.Fatal(err)
	}
	f, err := s.Files().ByName("big-test-file-150.dat")
	if err != nil {
		t.Fatal(err)
	}
	f2, err := s.Files().ByName("small-test-file-2.dat")
	if err != nil {
		t.Fatal(err)
	}
	return f, s, []interf.File{f, f2}
}

type testStat struct {
	t  *testing.T
	at interf.ReaderAt

	CacheHit      uint64
	CacheMis      uint64
	CacheSet      uint64
	RAtNew        uint64
	RAtClosing    uint64
	RAtClose      uint64
	RAtReq        uint64
	RAtRetErr     uint64
	RAtSectorSkip uint64
	RAtSectorRet  uint64
	RAtBest       uint64
	RAtAdd        uint64
	RAtAddErr     uint64
}

func (ts *testStat) Check() {
	m := ts.at.Stat()

	if m["RAtClosing"] != ts.RAtClosing {
		ts.t.Errorf("RAtClosing: should=%d, is=%d", ts.RAtClosing, m["RAtClosing"])
	}
	if m["RAtNew"] != ts.RAtNew {
		ts.t.Errorf("RAtNew: should=%d, is=%d", ts.RAtNew, m["RAtNew"])
	}
	if m["CacheSet"] != ts.CacheSet {
		ts.t.Errorf("CacheSet: should=%d, is=%d", ts.CacheSet, m["CacheSet"])
	}
	if m["CacheMis"] != ts.CacheMis {
		ts.t.Errorf("CacheMis: should=%d, is=%d", ts.CacheMis, m["CacheMis"])
	}
	if m["CacheHit"] != ts.CacheHit {
		ts.t.Errorf("CacheHit: should=%d, is=%d", ts.CacheHit, m["CacheHit"])
	}
	if m["RAtClose"] != ts.RAtClose {
		ts.t.Errorf("RAtClose: should=%d, is=%d", ts.RAtClose, m["RAtClose"])
	}
	if m["RAtReq"] != ts.RAtReq {
		ts.t.Errorf("RAtReq: should=%d, is=%d", ts.RAtReq, m["RAtReq"])
	}
	if m["RAtRetErr"] != ts.RAtRetErr {
		ts.t.Errorf("RAtRetErr: should=%d, is=%d", ts.RAtRetErr, m["RAtRetErr"])
	}
	if m["RAtSectorSkip"] != ts.RAtSectorSkip {
		ts.t.Errorf("RAtSectorSkip: should=%d, is=%d", ts.RAtSectorSkip, m["RAtSectorSkip"])
	}
	if m["RAtSectorRet"] != ts.RAtSectorRet {
		ts.t.Errorf("RAtSectorRet: should=%d, is=%d", ts.RAtSectorRet, m["RAtSectorRet"])
	}
	if m["RAtBest"] != ts.RAtBest {
		ts.t.Errorf("RAtBest: should=%d, is=%d", ts.RAtBest, m["RAtBest"])
	}
	if m["RAtAdd"] != ts.RAtAdd {
		ts.t.Errorf("RAtAdd: should=%d, is=%d", ts.RAtAdd, m["RAtAdd"])
	}
	if m["RAtAddErr"] != ts.RAtAddErr {
		ts.t.Errorf("RAtAddErr: should=%d, is=%d", ts.RAtAddErr, m["RAtAddErr"])
	}
}
