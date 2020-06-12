package gdrive

import (
	"log"
	"time"
)

// ParseTime get a RFC 3339 date-time string and return as a Unix time.
// input example: 2018-08-03T12:03:30.407Z
func ParseTime(s string) int64 {
	// parse string
	t := new(time.Time)
	err := t.UnmarshalText([]byte(s))

	// error handling -> return invalid date
	if err != nil {
		log.Printf("ERROR: %s/parseTime: can's parse timestring '%s': %v", packageName, s, err)
		return time.Now().Unix() - 4730000000 // -150 years
	}

	// OK
	return t.Unix()
}
