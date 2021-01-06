package util

import (
	"log"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")

	getConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")
	getConsoleWindow      = kernel32.NewProc("GetConsoleWindow")
	showWindow            = user32.NewProc("ShowWindow")
)

func HideConsole() {
	if hwnd, _, _ := getConsoleWindow.Call(); hwnd == 0 {
		return // no window
	} else {
		var pid uint32
		if n, _, err := getConsoleProcessList.Call(uintptr(unsafe.Pointer(&pid)), 1); n == 0 {
			log.Fatal(err)
		} else if n > 1 {
			return // not the only process
		}
		showWindow.Call(hwnd, 0) // SW_HIDE
	}
}
