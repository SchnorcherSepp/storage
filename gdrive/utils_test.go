package gdrive_test

import (
	"github.com/SchnorcherSepp/storage/gdrive"
	"testing"
)

func TestParseTime(t *testing.T) {
	// define tests cases
	tests := []struct {
		dateTime string
		unix     int64
	}{
		{"", -1},
		{"foo", -1},
		{"2002-10-02T10:00:00-05:00", 1033570800},
		{"2002-10-02T15:00:00Z", 1033570800},
		{"2002-10-02T15:00:00.05Z", 1033570800},
		{"1994-11-05T08:15:30-05:00", 784041330},
		{"1994-11-05T13:15:30Z", 784041330},
		{"2018-08-03T12:03:30.407Z", 1533297810},
	}

	// run test
	for i, a := range tests {
		v := gdrive.ParseTime(a.dateTime)
		// errors
		if v < 0 && a.unix < 0 {
			continue
		}
		// OKs
		if v != a.unix {
			t.Errorf("test case %d fail: RFC 3339 date-time string '%s' is %d and should %d", i+1, a.dateTime, v, a.unix)
		}
	}
}
