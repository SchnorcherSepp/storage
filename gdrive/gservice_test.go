package gdrive_test

import (
	"bytes"
	"fmt"
	impl "github.com/SchnorcherSepp/storage/defaultimpl"
	"github.com/SchnorcherSepp/storage/gdrive"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	google "google.golang.org/api/drive/v3"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
)

var (
	testIndexCacheFile = path.Join(os.TempDir(), "testIndexCacheFile.dat")
	testServiceRO      interf.Service
	testServiceWR      interf.Service
	testOauthRO        *google.Service
	testOauthWR        *google.Service
)

func init() {
	var err error

	// readonly oauth
	testOauthRO, err = gdrive.OAuth(testClientCredFile, testTokenFileRead, true)
	if err != nil {
		panic(err)
	}
	testServiceRO = gdrive.NewGService("root", testIndexCacheFile, testOauthRO, nil, true)

	// read/write oauth
	testOauthWR, err = gdrive.OAuth(testClientCredFile, testTokenFileWrite, false)
	if err != nil {
		panic(err)
	}
	testServiceWR = gdrive.NewGService("", testIndexCacheFile, testOauthWR, nil, true)

	// write demo files
	err = impl.InitDemo(testServiceWR)
	if err != nil {
		panic(err)
	}
}

func TestService_Save_Read_Trash(t *testing.T) {
	// test data
	testBytes := []byte("Test Bytes Foo Bar")
	testFileName := "TestService_Save_Read_Trash_File1"

	// wrong permissions (can't write with ReadOnly service)
	fileId, err := testServiceRO.Save(testFileName, bytes.NewReader(testBytes), 0)
	if err == nil {
		// no error ??
		_ = testServiceRO.Trash(fileId)
		t.Errorf("Save() was with readonly service successful! Recreate %s with readonly service", testTokenFileRead)
	}
	if !strings.Contains(fmt.Sprintf("%v", err), "wrong permissions") {
		// error must be "wrong permissions"
		t.Errorf("no 'wrong permissions' error: %v", err)
	}

	//--------------------------------------------------------------------

	// save file: name=testFileName, data=testBytes
	fileId, err = testServiceWR.Save(testFileName, bytes.NewReader(testBytes), 0)
	if err != nil {
		t.Error(err)
	}

	//--------------------------------------------------------------------

	// get LimitedReader with offset 1 and n=16 => don't read fist and last byte
	reader, err := testServiceRO.LimitedReader(fileId, 1, 16)
	if err != nil {
		t.Error(err)
	}
	b, err := ioutil.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "est Bytes Foo Ba" {
		t.Fatalf("read error: %s", b)
	}

	// get LimitedReader with offset 10 and n=0
	reader, err = testServiceRO.LimitedReader(fileId, 11, 0)
	if err != nil {
		t.Error(err)
	}
	b, err = ioutil.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "" {
		t.Fatalf("read error: %s", b)
	}

	// get LimitedReader with offset 10 and n=1
	reader, err = testServiceRO.LimitedReader(fileId, 11, 1)
	if err != nil {
		t.Error(err)
	}
	b, err = ioutil.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "F" {
		t.Fatalf("read error: %s", b)
	}

	//--------------------------------------------------------------------

	// Trash
	err = testServiceWR.Trash(fileId)
	if err != nil {
		t.Errorf("Trash() fail: %v", err)
	}

	//--------------------------------------------------------------------

	// Save() with LimitReader
	fileId, err = testServiceWR.Save(testFileName, bytes.NewReader(testBytes), 4)

	// read all
	reader, err = testServiceRO.Reader(fileId, 0)
	if err != nil {
		t.Error(err)
	}
	b, err = ioutil.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "Test" {
		t.Fatalf("read error: %s", b)
	}

	//--------------------------------------------------------------------

	// Trash
	err = testServiceWR.Trash(fileId)
	if err != nil {
		t.Error(err)
	}
}

func TestService_Update(t *testing.T) {
	// prepare
	indexCache := path.Join(os.TempDir(), "cacheSigTest.dat")
	_ = os.Remove(indexCache)
	cache := impl.NewCache(1)
	service := gdrive.NewGService("root", indexCache, testOauthWR, cache, true)

	/*
				For this test we have to check the log output.
				 1) buf := bytes.NewBufferString("") // new buffer
				 2) log.SetOutput(buf) // write logs to buffer
		         3) buf.String()
	*/

	// ----- test 1: Init without indexcache -------------------------------------
	buf := bytes.NewBufferString("") // new buffer
	log.SetOutput(buf)               // write logs to buffer

	err := service.Update()

	if err != nil || len(service.Files().All()) <= 1000 || !strings.Contains(buf.String(), "initialization without indexcache") {
		t.Fatalf("test fail")
	}

	// ----- test 2: Update ------------------------------------------------------
	buf = bytes.NewBufferString("") // new buffer
	log.SetOutput(buf)              // write logs to buffer

	err = service.Update()

	if err != nil || len(service.Files().All()) <= 1000 || !strings.Contains(buf.String(), "successful file update") {
		t.Fatalf("test fail")
	}

	// ----- test 3: Update with file changes ------------------------------------------------------
	buf = bytes.NewBufferString("") // new buffer
	log.SetOutput(buf)              // write logs to buffer

	f, err := service.Save("update.test", bytes.NewReader([]byte{}), 0)
	if err != nil {
		t.Fatal(err)
	}
	count1 := len(service.Files().All())
	err = service.Update()
	count2 := len(service.Files().All())
	if err != nil {
		t.Fatal(err)
	}
	err = service.Trash(f)
	if err != nil {
		t.Fatal(err)
	}

	if count1+1 != count2 {
		t.Fatalf("test fail: c1=%d, c2=%d", count1, count2)
	}

	// ----- test 4: Init WITH indexcache ----------------------------------------
	buf = bytes.NewBufferString("") // new buffer
	log.SetOutput(buf)              // write logs to buffer

	service = gdrive.NewGService("root", indexCache, testOauthWR, cache, true)
	err = service.Update()

	if err != nil || len(service.Files().All()) <= 1000 || !strings.Contains(buf.String(), "speed up initialization with indexcache") {
		t.Fatalf("test fail")
	}

	// ----- test 5: Init WITH indexcache BUT wrong cache sig -----------------------
	buf = bytes.NewBufferString("") // new buffer
	log.SetOutput(buf)              // write logs to buffer

	service = gdrive.NewGService("xXx", indexCache, testOauthRO, cache, true) // change parent folder to invalidate cacheSig
	_ = service.Update()                                                      // ignore errors

	if !strings.Contains(buf.String(), "wrong indexcache signature") {
		t.Fatalf("test fail")
	}

	// reset logger
	log.SetOutput(os.Stdout)
	log.Printf("reset fin")
}

//--------------------------------------------------------------------------------------------------------------------//

func TestRace_Update(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(5)
	for n := 0; n < 5; n++ {
		go func() {
			//------------------------------
			for i := 0; i < 6; i++ {
				err := testServiceWR.Update()
				if err != nil {
					t.Fatal(err)
				}
			}
			//------------------------------
			wg.Done()
		}()
	}
	wg.Wait()
}
