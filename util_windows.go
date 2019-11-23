package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

//go:generate goversioninfo -64 -icon=assets/favicon.ico -manifest=win.manifest

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

func exitChrome(cmd *exec.Cmd) {
	for i := 0; i < 10; i++ {
		if exec.Command("taskkill", "/pid", strconv.Itoa(cmd.Process.Pid)).Run() != nil {
			return
		}
		time.Sleep(time.Second / 10)
	}
}

func openURLCmd(url string) *exec.Cmd {
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url)
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

func getANSIPath(path string) (string, error) {
	path = filepath.Clean(path)
	vol := len(filepath.VolumeName(path))

	if vol > 2 {
		return "", errors.New("UNC path not supported: " + path)
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getShortPathName := kernel32.NewProc("GetShortPathNameW")
	wideCharToMultiByte := kernel32.NewProc("WideCharToMultiByte")

	sep := len(path)
	for {
		_, err := os.Stat(path[:sep])
		if os.IsNotExist(err) {
			i := sep - 1
			for i >= vol && !os.IsPathSeparator(path[i]) {
				i--
			}
			if i >= vol {
				sep = i
				continue
			}
		}
		if err == nil {
			file := path[:sep]
			if filepath.IsAbs(file) {
				file = `\\?\` + file
			}
			if long, err := syscall.UTF16FromString(file); err == nil {
				short := [264]uint16{}
				n, _, _ := getShortPathName.Call(
					uintptr(unsafe.Pointer(&long[0])),
					uintptr(unsafe.Pointer(&short[0])), 264)
				if 0 < n && n < 264 {
					file = syscall.UTF16ToString(short[:n])
					path = strings.TrimPrefix(file, `\\?\`) + path[sep:]
				}
			}
		}
		break
	}

	if long, err := syscall.UTF16FromString(path); err == nil {
		var used int32
		n, _, _ := wideCharToMultiByte.Call(0 /*CP_ACP*/, 0x400, /*WC_NO_BEST_FIT_CHARS*/
			uintptr(unsafe.Pointer(&long[0])), ^uintptr(0), 0, 0, 0,
			uintptr(unsafe.Pointer(&used)))

		if 0 < n && n < 260 && used == 0 {
			return path, nil
		}
	}

	return path, errors.New("Could not convert to ANSI path: " + path)
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

	n, _, err := setConsoleCtrlHandler.Call(
		syscall.NewCallback(func(controlType uint) uint {
			if controlType >= 2 {
				c <- syscall.Signal(0x1f + controlType)
				select {}
			}
			return 0
		}),
		1)

	if n == 0 {
		log.Fatal(err)
	}
}
