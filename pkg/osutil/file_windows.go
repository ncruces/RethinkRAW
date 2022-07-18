package osutil

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

func isHidden(fi os.FileInfo) bool {
	s, ok := fi.Sys().(*syscall.Win32FileAttributeData)
	return ok && s.FileAttributes&(syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_SYSTEM) != 0
}

func open(file string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", file).Run()
}

func isANSIString(s string) bool {
	if s == "" {
		return true
	}

	var used int32
	long := utf16.Encode([]rune(s))
	n, _, _ := wideCharToMultiByte.Call(0 /*CP_ACP*/, 0x400, /*WC_NO_BEST_FIT_CHARS*/
		uintptr(unsafe.Pointer(&long[0])), uintptr(len(long)), 0, 0, 0,
		uintptr(unsafe.Pointer(&used)))

	return n > 0 && used == 0
}

func getANSIPath(path string) (string, error) {
	path = filepath.Clean(path)

	if len(path) < 260 && isANSIString(path) {
		return path, nil
	}

	vol := len(filepath.VolumeName(path))
	for i := len(path); i >= vol; i-- {
		if i == len(path) || os.IsPathSeparator(path[i]) {
			file := path[:i]
			_, err := os.Stat(file)
			if err == nil {
				if filepath.IsAbs(file) {
					file = `\\?\` + file
				}
				if long, err := syscall.UTF16FromString(file); err == nil {
					short := [264]uint16{}
					n, _ := windows.GetShortPathName(&long[0], &short[0], 264)
					if 0 < n && n < 264 {
						file = syscall.UTF16ToString(short[:n])
						path = strings.TrimPrefix(file, `\\?\`) + path[i:]
						if len(path) < 260 && isANSIString(path) {
							return path, nil
						}
					}
				}
				break
			}
		}
	}

	return path, errors.New("could not convert to ANSI path: " + path)
}
