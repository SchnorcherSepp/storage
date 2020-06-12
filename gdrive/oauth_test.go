package gdrive_test

import (
	"github.com/SchnorcherSepp/storage/gdrive"
	"os"
	"path"
	"strings"
	"testing"
)

const testClientCredFile = "../test/secret/client_credentials.json"
const testTokenFileRead = "../test/secret/token_read.json"
const testTokenFileWrite = "../test/secret/token_write.json"

func TestOAuth_secret(t *testing.T) {
	_, err := gdrive.OAuth(testClientCredFile, testTokenFileRead, false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gdrive.OAuth(testClientCredFile, testTokenFileWrite, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOAuth(t *testing.T) {
	// loadOAuthConf: read file error (not exist?)
	s, e := gdrive.OAuth("", "", false)
	if s != nil || e == nil || !strings.Contains(e.Error(), "cannot find the file") {
		t.Errorf("s=%v, e=%v", s, e)
	}

	// loadOAuthConf: parsing error (empty file?)
	s, e = gdrive.OAuth(emptyFile(), "", false)
	if s != nil || e == nil || !strings.Contains(e.Error(), "unexpected end of JSON input") {
		t.Errorf("s=%v, e=%v", s, e)
	}

	// can't test this (user interaction)
	// * loadToken: open error (not exist?)
	// * loadToken: parsing error (empty file?)
	// * reqNewToken: request (new)
	// * reqNewToken: scope (default: read & write access)
	var _ = "nop"

	// valid clientCred and valid token (READ)
	s, e = gdrive.OAuth("../test/secret/client_credentials.json", "../test/secret/token_read.json", true)
	if e != nil || s == nil {
		t.Errorf("s=%v, e=%v", s, e)
	}

	// valid clientCred and valid token (WRITE)
	s, e = gdrive.OAuth("../test/secret/client_credentials.json", "../test/secret/token_write.json", false)
	if e != nil || s == nil {
		t.Errorf("s=%v, e=%v", s, e)
	}
}

//--------  HELPER  --------------------------------------------------------------------------------------------------//

func emptyFile() string {
	p := path.Join(os.TempDir(), "empty.file")

	fh, err := os.Create(p)
	if err != nil {
		panic(err)
	}
	_ = fh.Close()

	return p
}
