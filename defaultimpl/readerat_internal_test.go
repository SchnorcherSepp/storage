package impl

import (
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"github.com/oxtoacart/bpool"
	"io/ioutil"
	"sync"
	"testing"
)

func Test_bestConn(t *testing.T) {
	r := _ReaderAt{
		inner: make([]*_Reader, interf.MaxReadersPerFile),
		stat:  new(_ReaderStat),
		file:  NewFile("fileId", "name.file", 1234, 5678, "x0x0x0x0x0x0x0"),
	}

	r.inner[0] = newInnerReader(ioutil.NopCloser(nil), 22000) // 3
	r.inner[1] = newInnerReader(ioutil.NopCloser(nil), 33000) // 4
	r.inner[2] = newInnerReader(ioutil.NopCloser(nil), 0)     // 1
	r.inner[3] = newInnerReader(ioutil.NopCloser(nil), 44000) // 5
	r.inner[4] = newInnerReader(ioutil.NopCloser(nil), 99000) // 6
	r.inner[5] = newInnerReader(ioutil.NopCloser(nil), 11000) // 2

	// exact
	if c := r.bestConn(22000); c != r.inner[0] {
		t.Errorf("wromng reader: %#v", r)
	}
	if c := r.bestConn(33000); c != r.inner[1] {
		t.Errorf("wromng reader: %#v", c)
	}
	if c := r.bestConn(0); c != r.inner[2] {
		t.Errorf("wromng reader: %#v", c)
	}
	if c := r.bestConn(44000); c != r.inner[3] {
		t.Errorf("wromng reader: %#v", c)
	}
	if c := r.bestConn(99000); c != r.inner[4] {
		t.Errorf("wromng reader: %#v", c)
	}
	if c := r.bestConn(11000); c != r.inner[5] {
		t.Errorf("wromng reader: %#v", c)
	}

	// small jump
	if c := r.bestConn(7); c != r.inner[2] {
		t.Errorf("wromng reader: %#v", c)
	}
	if c := r.bestConn(33000 + interf.MaxSectorJump); c != r.inner[1] {
		t.Errorf("wromng reader: %#v", c)
	}
	if c := r.bestConn(99001); c != r.inner[4] {
		t.Errorf("wromng reader: %#v", c)
	}

	// invalid
	if c := r.bestConn(22000 - 1); c != nil {
		t.Errorf("wromng reader: %#v", c)
	}
	if c := r.bestConn(33000 + interf.MaxSectorJump + 1); c != nil {
		t.Errorf("wromng reader: %#v", c)
	}

	// no reader
	r.inner = make([]*_Reader, interf.MaxReadersPerFile)
	if c := r.bestConn(0); c != nil {
		t.Errorf("wromng reader: %#v", c)
	}
}

func Test_addConn(t *testing.T) {
	// init ram service
	s := NewRamService(nil, false)
	if err := InitDemo(s); err != nil {
		t.Fatal(err)
	}
	f, err := s.Files().ByName("big-test-file-150.dat")
	if err != nil {
		t.Fatal(err)
	}

	r := &_ReaderAt{
		mux:     new(sync.Mutex),
		inner:   make([]*_Reader, interf.MaxReadersPerFile),
		stat:    new(_ReaderStat),
		file:    f,
		service: s,
		cache:   nil,
		pool:    bpool.NewBytePool(25, interf.SectorSize),
	}

	for i := 0; i < interf.MaxReadersPerFile; i++ {
		// new
		c, err := r.addConn(uint64(i))
		if err != nil {
			t.Errorf("%v", err)
		}

		// check return
		if c != r.inner[0] {
			t.Errorf("wrong position")
		}

		// check list
		nilCount := 0
		initCount := 0
		for _, v := range r.inner {
			if v == nil {
				nilCount++
			} else {
				initCount++
			}
		}
		if initCount != i+1 || nilCount != interf.MaxReadersPerFile-i-1 {
			t.Errorf("error: initCount=%d, nilCount=%d", initCount, nilCount)
		}
	}

	// check reader order
	for i := 1; i < len(r.inner); i++ {
		if r.inner[i-1].sector <= r.inner[i].sector {
			t.Errorf("wrong order: %d", i)
		}
	}

	// add new reader to full list
	_, err = r.addConn(uint64(99))
	if err != nil {
		t.Errorf("%v", err)
	}
	last := len(r.inner) - 1
	if r.inner[last].sector != 1 || r.inner[0].sector != 99 {
		t.Errorf("final error")
	}
}

func Test_sortByAge(t *testing.T) {
	r := _ReaderAt{
		inner: make([]*_Reader, interf.MaxReadersPerFile),
	}

	r.inner[0] = &_Reader{age: 1, c: ioutil.NopCloser(nil)}
	r.inner[1] = &_Reader{age: 2, c: ioutil.NopCloser(nil)}
	r.inner[2] = &_Reader{age: 0, c: ioutil.NopCloser(nil)}
	r.inner[3] = &_Reader{age: 3, c: ioutil.NopCloser(nil)}
	r.inner[4] = &_Reader{age: 5, c: ioutil.NopCloser(nil)}
	r.inner[5] = &_Reader{age: 4, c: ioutil.NopCloser(nil)}

	// sort
	r.sortByAge()
	arr := r.inner
	if arr[0].age != 5 || arr[1].age != 4 || arr[2].age != 3 || arr[3].age != 2 || arr[4].age != 1 || arr[5].age != 0 {
		t.Errorf("wrong order")
	}

	// invalid bottom
	r.inner[0] = nil
	r.inner[1].c = nil

	r.sortByAge()
	arr = r.inner
	if arr[0].age != 3 || arr[1].age != 2 || arr[2].age != 1 || arr[3].age != 0 {
		t.Errorf("wrong order")
	}
}

func Test_calcSector(t *testing.T) {
	r := _ReaderAt{}

	// test 1
	for sector := 0; sector < 50; sector++ {
		// request bytes
		for rb := 0; rb < interf.SectorSize; rb++ {
			calcId, innerOff := r.calcSector(int64(sector)*interf.SectorSize + int64(rb))
			// test calcId
			if calcId != uint64(sector) {
				t.Errorf("sector %d and request byte %d -> calcId = %d", sector, rb, calcId)
			}
			// test innerOff
			if innerOff != rb {
				t.Errorf("sector %d and request byte %d -> innerOff = %d", sector, rb, innerOff)
			}
		}
	}

	// test 2: invalid
	if calcId, innerOff := r.calcSector(-1); calcId != 0 || innerOff != 0 {
		t.Fatalf("invalid args test fail")
	}
}
