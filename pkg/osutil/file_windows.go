// Copyright (c) 2020 Nuno Cruces
// SPDX-License-Identifier: MIT

package osutil

import (
	"os"
	"os/exec"
	"syscall"
)

func isHidden(fi os.FileInfo) bool {
	s, ok := fi.Sys().(*syscall.Win32FileAttributeData)
	return ok && s.FileAttributes&(syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_SYSTEM) != 0
}

func open(file string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", file).Run()
}
