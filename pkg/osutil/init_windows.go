package osutil

import (
	"syscall"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")

	getShortPathName    = kernel32.NewProc("GetShortPathNameW")
	wideCharToMultiByte = kernel32.NewProc("WideCharToMultiByte")
	getConsoleWindow    = kernel32.NewProc("GetConsoleWindow")
	attachConsole       = kernel32.NewProc("AttachConsole")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
)
