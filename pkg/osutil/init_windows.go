package osutil

import (
	"syscall"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	getShortPathName    = kernel32.NewProc("GetShortPathNameW")
	wideCharToMultiByte = kernel32.NewProc("WideCharToMultiByte")
)
