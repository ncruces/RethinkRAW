package main

import (
	"crypto/md5"
	"encoding/base64"
	"mime"
	"sort"
	"strconv"
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

func hash(data string) string {
	h := md5.Sum([]byte(data))
	return base64.URLEncoding.EncodeToString(h[:15])
}

func index(a []string, x string) int {
	for k, v := range a {
		if x == v {
			return k
		}
	}
	return -1
}

func unique(a *[]string) {
	s := *a
	if len(s) < 1 {
		return
	}

	sort.Strings(s)

	i := 0
	for j := 1; j < len(s); j++ {
		if s[i] != s[j] {
			i++
			s[i] = s[j]
		}
	}
	i++

	*a = s[:i:i]
}

func toASCII(str string) string {
	builder := strings.Builder{}
	for _, r := range str {
		// control
		if r <= 0x1f || 0x7f <= r && r <= 0x9f {
			continue
		}
		// unicode
		if r >= 0xa0 {
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
		switch r {
		case '\\', '/', ':', '*', '?', '<', '>', '|':
			// Windows doesn't like these.
		case '"':
			builder.WriteByte('\'')
		case '.':
			builder.WriteByte('.')
			dots += 1
		default:
			if strconv.IsPrint(r) {
				builder.WriteRune(r)
			}
		}
	}

	if builder.Len() > dots {
		return builder.String()
	}
	return ""
}
