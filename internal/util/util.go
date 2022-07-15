package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
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
