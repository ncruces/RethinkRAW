package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"unsafe"
)

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func HashedID(data string) string {
	buf := md5.Sum([]byte(data))
	return base64.RawURLEncoding.EncodeToString(buf[:15])
}

func RandomID() string {
	var buf [15]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buf[:])
}

func PercentEncode(s string) string {
	const upperhex = "0123456789ABCDEF"
	unreserved := func(c byte) bool {
		switch {
		case c >= 'a':
			return c <= 'z' || c == '~'
		case c >= 'A':
			return c <= 'Z' || c == '_'
		case c >= '0':
			return c <= '9'
		default:
			return c == '-' || c == '.'
		}
	}

	hex := 0
	for _, c := range []byte(s) {
		if !unreserved(c) {
			hex++
		}
	}
	if hex == 0 {
		return s
	}

	i := 0
	buf := make([]byte, len(s)+2*hex)
	for _, c := range []byte(s) {
		if unreserved(c) {
			buf[i] = c
			i++
		} else {
			buf[i+0] = '%'
			buf[i+1] = upperhex[c>>4]
			buf[i+2] = upperhex[c&15]
			i += 3
		}
	}
	return *(*string)(unsafe.Pointer(&buf))
}
