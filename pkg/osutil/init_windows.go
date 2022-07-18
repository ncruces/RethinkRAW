package osutil

import "golang.org/x/sys/windows"

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	user32   = windows.NewLazySystemDLL("user32.dll")

	wideCharToMultiByte = kernel32.NewProc("WideCharToMultiByte")
	getConsoleWindow    = kernel32.NewProc("GetConsoleWindow")
	attachConsole       = kernel32.NewProc("AttachConsole")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
)
