package main

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"mime"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
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

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, in)
	return
}

func moveFile(src, dst string) error {
	err := os.Rename(src, dst)
	le, ok := err.(*os.LinkError)

	// 0x12 is EXDEV, 0x11 is ERROR_NOT_SAME_DEVICE
	if ok && (le.Err == syscall.Errno(0x12) || (le.Err == syscall.Errno(0x11) && runtime.GOOS == "windows")) {
		if err := copyFile(src, dst); err != nil {
			return err
		}
		if err := os.Remove(src); os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	return err
}

func lnkyFile(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}

	dfi, err := os.Stat(dst)
	if os.SameFile(sfi, dfi) {
		return nil
	}

	err = os.Link(src, dst)
	le, ok := err.(*os.LinkError)

	// 0x12 is EXDEV, 0x11 is ERROR_NOT_SAME_DEVICE
	if ok && (os.IsExist(le) || le.Err == syscall.Errno(0x12) || (le.Err == syscall.Errno(0x11) && runtime.GOOS == "windows")) {
		return copyFile(src, dst)
	}
	return err
}

func setupDirs() (err error) {
	if exe, err := os.Executable(); err != nil {
		return err
	} else {
		baseDir = filepath.Dir(exe)
	}

	dataDir = filepath.Join(baseDir, "data")
	tempDir = filepath.Join(os.TempDir(), "RethinkRAW")

	tempDir, err = getANSIPath(tempDir)
	if err != nil {
		return
	}

	testDir := func() error {
		if err := os.MkdirAll(dataDir, 0700); err != nil {
			return err
		}
		if f, err := os.Create(filepath.Join(dataDir, "lastrun")); err != nil {
			return err
		} else {
			return f.Close()
		}
	}
	if testDir() == nil {
		return
	}
	if data, err := os.UserConfigDir(); err != nil {
		return err
	} else {
		dataDir = filepath.Join(data, "RethinkRAW")
	}
	return testDir()
}
