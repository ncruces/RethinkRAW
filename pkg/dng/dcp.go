package dng

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
)

// GetDCPProfileName extracts the profile name from a DCP file.
func GetDCPProfileName(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(data) < 8 {
		return "", errors.New("dcp: could not read dcp header")
	}

	var endian binary.ByteOrder
	switch string(data[0:4]) {
	case "IIRC":
		endian = binary.LittleEndian
	case "MMCR":
		endian = binary.BigEndian
	default:
		return "", errors.New("dcp: could not read dcp header")
	}

	offset := endian.Uint32(data[4:])
	if len(data) < int(offset)+2 {
		return "", errors.New("dcp: invalid offset")
	}

	count := endian.Uint16(data[offset:])
	entries := data[offset+2:]
	if len(data) < 12*int(count) {
		return "", errors.New("dcp: invalid directory size")
	} else {
		entries = entries[:12*count]
	}

	for i := 0; i < len(entries); i += 12 {
		tag := endian.Uint16(entries[i:])
		if tag == 0xc6f8 { // ProfileName
			typ := endian.Uint16(entries[i+2:]) // BYTE or ASCII
			cnt := endian.Uint32(entries[i+4:])
			off := endian.Uint32(entries[i+8:])

			if (typ == 1 || typ == 2) && cnt > 1 {
				var val []byte
				if cnt <= 4 {
					val = entries[i+8:][:cnt]
				} else if len(data) >= int(off+cnt) {
					val = data[off:][:cnt]
				} else {
					return "", errors.New("dcp: invalid offset")
				}
				if bytes.IndexByte(val, 0) != int(cnt-1) { // NUL terminator
					return "", errors.New("dcp: invalid profile name")
				}
				return string(val[:cnt-1]), nil
			}
			return "", errors.New("dcp: invalid profile name")
		}
	}

	return "", errors.New("dcp: no profile name")
}
