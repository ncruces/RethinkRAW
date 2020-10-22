package exiftool

import (
	"bytes"
	"errors"
)

// Unmarshal parses the standard, short or veryShort ExifTool output formats.
// Loads tag names and values into a map.
func Unmarshal(data []byte, m map[string][]byte) error {
	for len(data) > 0 {
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			return errors.New("exiftool: unexpected end of output")
		}

		j := bytes.Index(data[:i], []byte(": "))
		if j < 0 {
			return errors.New("exiftool: missing separator")
		}

		key := bytes.TrimSpace(data[:j])
		val := bytes.TrimSuffix(data[j+2:i], []byte("\r"))
		m[string(key)] = val
		data = data[i+1:]
	}
	return nil
}
