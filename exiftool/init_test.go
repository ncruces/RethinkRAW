package exiftool

import (
	"runtime"
)

const path = "dist/exiftool"

var arg1 string

func init() {
	if runtime.GOOS == "windows" {
		arg1 = path
	}
}
