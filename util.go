package main

import (
	"crypto/md5"
	"encoding/base64"
)

const MaxUint = ^uint(0)
const MinUint = 0
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1

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
