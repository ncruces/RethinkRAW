// Package craw provides support for loading Camera Raw settings.
package craw

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// IndexedRecord holds properties for an indexed file.
type IndexedRecord struct {
	Path string            // The path to the indexed file.
	Prop map[string]string // The name-value property pairs.
}

// LoadIndex loads an Index.dat file.
// Index.dat files index the many profiles, presets and other settings
// in a Camera Raw settings directory for faster access.
func LoadIndex(path string) ([]IndexedRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Index.dat files are a collection of records.
	//
	// There's an 8 byte file header:
	//  - a 4 byte number, presumably the index's version
	//  - a 4 byte number, the number of records in the index
	//
	// Then, follow N records:
	//  - a string, the file path to which the record refers
	//  - a 8 byte record header, contents unknown
	//  - a 4 byte number, the number of properties for the record, and
	//  - 2*N strings, N name-value property pairs
	//
	// Numbers are stored in little endian.
	// Strings are stored with a 4 byte length, followed by N bytes of content,
	// and a null terminator byte.

	var buf [12]byte

	// Read the 8 byte file header
	if _, err := io.ReadFull(f, buf[:8]); err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	count := int(binary.LittleEndian.Uint32(buf[4:]))
	index := make([]IndexedRecord, count)

	for i := 0; i < count; i++ {
		index[i].Path, err = readString(f)
		if err != nil {
			return index, err
		}

		index[i].Path = fixPath(index[i].Path)

		// Read the 12 byte record header
		if _, err := io.ReadFull(f, buf[:12]); err != nil {
			return index, err
		}

		count := int(binary.LittleEndian.Uint32(buf[8:]))
		index[i].Prop = make(map[string]string, count)

		for j := 0; j < count; j++ {
			key, err := readString(f)
			if err != nil {
				return index, err
			}

			val, err := readString(f)
			if err != nil {
				return index, err
			}

			index[i].Prop[key] = val
		}
	}

	return index, nil
}

func readString(r io.Reader) (string, error) {
	var buf [4]byte

	// Read the string length
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return "", err
	}

	len := int(binary.LittleEndian.Uint32(buf[:]))

	// Read the string with null terminator
	var str = make([]byte, len+1)
	if _, err := io.ReadFull(r, str); err != nil {
		return "", err
	}
	if str[len] != 0 {
		return "", errors.New("string not null terminated")
	}

	return string(str[:len]), nil
}
