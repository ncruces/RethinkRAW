package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"
)

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1

var MinTime = time.Unix(-2208988800, 0) // Jan 1, 1900
var MaxTime = MinTime.Add(1<<63 - 1)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func Must(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}

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

func Index(a []string, s string) int {
	for k, v := range a {
		if s == v {
			return k
		}
	}
	return -1
}

func Contains(a []string, s string) bool {
	return Index(a, s) >= 0
}

func LogPretty(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		log.Println(string(b))
	}
	return
}
