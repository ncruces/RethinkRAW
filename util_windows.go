package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func openURLCmd(url string) *exec.Cmd {
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url)
}

func findChrome() string {
	versions := []string{`Google\Chrome`, `Google\Chrome SxS`, "Chromium"}
	prefixes := []string{os.Getenv("LOCALAPPDATA"), os.Getenv("PROGRAMFILES"), os.Getenv("PROGRAMFILES(X86)")}
	suffix := `\Application\chrome.exe`

	for _, v := range versions {
		for _, p := range prefixes {
			if p != "" {
				c := filepath.Join(p, v, suffix)
				if _, err := os.Stat(c); err == nil {
					return c
				}
			}
		}
	}

	return ""
}

func isHidden(fi os.FileInfo) bool {
	if strings.HasPrefix(fi.Name(), ".") {
		return true
	}

	if s, ok := fi.Sys().(*syscall.Win32FileAttributeData); ok &&
		s.FileAttributes&(syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_SYSTEM) != 0 {
		return true
	}

	return false
}

func hideConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")

	getConsoleProcessList := kernel32.NewProc("GetConsoleProcessList")
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	showWindow := user32.NewProc("ShowWindow")

	var pid uint32
	if n, _, err := getConsoleProcessList.Call(uintptr(unsafe.Pointer(&pid)), 1); n == 0 {
		log.Fatal(err)
	} else if n > 1 {
		return // not the only process
	}

	if hwnd, _, _ := getConsoleWindow.Call(); hwnd == 0 {
		return // no window
	} else {
		showWindow.Call(hwnd, 0) // SW_HIDE
	}
}

func bringToTop() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")

	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	setForegroundWindow := user32.NewProc("SetForegroundWindow")

	if hwnd, _, _ := getConsoleWindow.Call(); hwnd == 0 {
		return // no window
	} else {
		setForegroundWindow.Call(hwnd)
	}
}

func handleConsoleCtrl(c chan<- os.Signal) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	setConsoleCtrlHandler := kernel32.NewProc("SetConsoleCtrlHandler")

	if n, _, err := setConsoleCtrlHandler.Call(
		syscall.NewCallback(func(controlType uint) uint {
			if controlType >= 2 {
				c <- syscall.Signal(0x1f + controlType)
				time.Sleep(30 * time.Second)
			}
			return 0
		}),
		1); n == 0 {
		log.Fatal(err)
	}
}
