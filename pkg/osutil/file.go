// Package osutil provides additional platform-independent access to operating system functionality.
package osutil

import (
	"errors"
	"io"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var newFilenameRE = regexp.MustCompile(`\A(.*?)(?: \((\d{1,4})\))?(\.\w*)?\z`)

// NewFile creates a new named file.
// If the file already exists, a numeric suffix is appended or incremented.
func NewFile(name string) (*os.File, error) {
	for {
		f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if os.IsExist(err) {
			m := newFilenameRE.FindStringSubmatch(name)
			if m != nil {
				var i = 0
				if m[2] != "" {
					i, _ = strconv.Atoi(m[2])
				}
				name = m[1] + " (" + strconv.Itoa(i+1) + ")" + m[3]
				continue
			}
		}
		return f, err
	}
}

// Copy copies src to dst.
func Copy(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, in)
	return err
}

// Move moves src to dst.
// Tries os.Rename. Failing that, does a Copy followed by a os.Remove.
func Move(src, dst string) error {
	err := os.Rename(src, dst)
	if isNotSameDevice(err) {
		if err := Copy(src, dst); err != nil {
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

// Lnky copies src to dst.
// Tries os.Link to create a hardlink. Failing that, does a Copy.
func Lnky(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}

	dfi, _ := os.Stat(dst)
	if os.SameFile(sfi, dfi) {
		return nil
	}

	err = os.Link(src, dst)
	if os.IsExist(err) || isNotSameDevice(err) {
		return Copy(src, dst)
	}
	return err
}

func isNotSameDevice(err error) bool {
	var lerr *os.LinkError
	if errors.As(err, &lerr) {
		if runtime.GOOS == "windows" {
			return lerr.Err == syscall.Errno(0x11) // ERROR_NOT_SAME_DEVICE
		} else {
			return lerr.Err == syscall.Errno(0x12) // EXDEV
		}
	}
	return false
}

// HiddenFile reports whether fi is hidden.
// Files starting with a period are reported as hidden on all systems, even Windows.
// Other than that, plaform rules apply.
func HiddenFile(fi os.FileInfo) bool {
	if strings.HasPrefix(fi.Name(), ".") {
		return true
	}
	return isHidden(fi)
}

// ShellOpen opens a file (or a directory, or URL),
// just as if you had double-clicked the file's icon.
func ShellOpen(file string) error {
	return open(file)
}
