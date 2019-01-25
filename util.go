package main

import (
	"crypto/md5"
	"encoding/base64"
	"mime"
	"strings"
)

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1

type constError string

func (e constError) Error() string { return string(e) }

func init() {
	must(mime.AddExtensionType(".dng", "image/x-adobe-dng"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func md5sum(data string) string {
	h := md5.Sum([]byte(data))
	return base64.URLEncoding.EncodeToString(h[:15])
}

func toASCII(str string) string {
	builder := strings.Builder{}
	for _, r := range str {
		// control
		if r < ' ' || 0x7f <= r && r < 0xa0 {
			continue
		}
		// unicode
		if r >= 0x7f {
			builder.WriteByte('?')
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func filename(name string) string {
	builder := strings.Builder{}
	dots := 0

	for _, r := range name {
		// control
		if r < ' ' || 0x7f <= r && r < 0xa0 {
			continue
		}
		switch r {
		// invalid
		case '\\', '/', ':', '*', '?', '<', '>', '|':
			continue
		case '"':
			builder.WriteByte('\'')
		case '.':
			builder.WriteByte('.')
			dots += 1
		default:
			builder.WriteRune(r)
		}
	}

	if builder.Len() > dots {
		return builder.String()
	}
	return ""
}
