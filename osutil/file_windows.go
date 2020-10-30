package osutil

import (
	"os"
	"syscall"
)

func isHidden(fi os.FileInfo) bool {
	s, ok := fi.Sys().(*syscall.Win32FileAttributeData)
	return ok && s.FileAttributes&(syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_SYSTEM) != 0
}
