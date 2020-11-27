package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"mime"
	"sort"
	"strconv"
	"strings"
)

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1

func init() {
	Check(mime.AddExtensionType(".dng", "image/x-adobe-dng"))
}

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

func Index(a []string, x string) int {
	for k, v := range a {
		if x == v {
			return k
		}
	}
	return -1
}

func Unique(a *[]string) {
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

func ToASCII(str string) string {
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

func Filename(name string) string {
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

func LogPretty(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		log.Println(string(b))
	}
	return
}
