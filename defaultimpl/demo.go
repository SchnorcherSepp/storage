package impl

import (
	"bytes"
	"fmt"
	interf "github.com/SchnorcherSepp/storage/interfaces"
	"math/rand"
	"strings"
)

// InitDemo creates the following test files:
//
//  + over 1000 small test files
//     Name: small-test-file-%d.dat
//     Data: text == filename
//  + 150 MB big test file (150*1024*1024 + 1 byte)
//     Name: big-test-file-150.dat
//     Data: random bytes
//     MD5:  4DB84342E76FC08B993C528E2C2CDC6A
//  + ...
//
func InitDemo(s interf.Service) error {

	// first, update file index
	err := s.Update()
	if err != nil {
		panic(err)
	}

	// create over 1000 small test files for Files() tests (if not exist)
	for i := 1; i < 1011; i++ {
		name := fmt.Sprintf("small-test-file-%d.dat", i)
		_, err := s.Files().ByName(name)
		if err != nil {
			_, err := s.Save(name, strings.NewReader(name), 0)
			if err != nil {
				return err
			}
		}
	}

	// create big random test files
	rnd := rand.New(rand.NewSource(1337))

	name := "big-test-file-150.dat"
	size := 150*1024*1024 + 1
	_, err = s.Files().ByName(name)
	if err != nil {
		_, err := s.Save(name, rnd, int64(size))
		if err != nil {
			return err
		}
	}

	// create type test files (like text, image, ...)
	const fuse = 128 * 1024         // 128 kB
	const buffer = 16 * 1024 * 1024 // 16 MB
	const comp = 1 * 1024 * 1024    // 1 MB
	const bundle = 12 * 1024 * 1024 // 12 MB

	for _, size := range []int{fuse, fuse - 1, fuse + 1, buffer, buffer - 1, buffer + 1, comp, comp - 1, comp + 1, bundle, bundle - 1, bundle + 1} {
		for _, rate := range []float32{0.99, 0.66, 0.33} {
			// data
			data := make([]byte, size)
			rnd.Read(data)
			for i := 0; i < int(float32(size)*rate); i++ {
				data[i] = 'B'
			}
			data[0] = 'A'

			// save
			name := fmt.Sprintf("special-file-%d-%f.dat", size, rate)
			_, err = s.Files().ByName(name)
			if err != nil {
				_, err := s.Save(name, bytes.NewReader(data), 0)
				if err != nil {
					return err
				}
			}
		}
	}

	// final update
	err = s.Update()
	if err != nil {
		panic(err)
	}
	return nil
}
