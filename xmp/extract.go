// Package xmp provides support for extracting XMP packets from data.
package xmp

import (
	"bufio"
	"bytes"
	"io"
)

const xpacket_begin = "<?xpacket begin="
const xpacket_end = "<?xpacket end="
const xpacket_sufix = `"w"?>`
const xpacket_id = "W5M0MpCehiHzreSzNTczkc9d"

func splitPacket(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if begin := bytes.Index(data, []byte(xpacket_begin)); begin >= 0 {
		if end := bytes.Index(data[begin+len(xpacket_begin):], []byte(xpacket_end)); end >= 0 {
			if last := begin + len(xpacket_begin) + end + len(xpacket_end) + len(xpacket_sufix); last < len(data) {
				if bytes.Contains(data[begin:begin+50], []byte(xpacket_id)) {
					return last, data[begin:last], nil
				}
			}
		}
		advance = begin
	} else {
		advance = len(data) - len(xpacket_begin) + 1
	}

	if atEOF {
		return 0, nil, io.EOF
	}
	return advance, nil, nil
}

// ExtractXMP extracts a XMP packet from the reader.
func ExtractXMP(r io.Reader) ([]byte, error) {
	scan := bufio.NewScanner(r)
	scan.Split(splitPacket)
	if scan.Scan() {
		return scan.Bytes(), nil
	}
	return nil, scan.Err()
}
