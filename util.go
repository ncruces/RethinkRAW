package main

import (
	"crypto/md5"
	"encoding/base64"
	"log"
	"strings"
	"syscall"
	"unsafe"
)

const MaxUint = ^uint(0)
const MinUint = 0
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1

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

func filename(name string) string {
	if name == "" {
		return ""
	}

	dots := 0
	builder := strings.Builder{}
	for _, r := range name {
		if r < ' ' {
			continue
		}
		switch r {
		case 0x7f, '\\', '/', ':', '*', '?', '<', '>', '|':
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

func hideConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")

	getConsoleProcessList := kernel32.NewProc("GetConsoleProcessList")
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	showWindow := user32.NewProc("ShowWindow")
	if err := getConsoleProcessList.Find(); err != nil {
		log.Fatal(err)
	}
	if err := getConsoleWindow.Find(); err != nil {
		log.Fatal(err)
	}
	if err := showWindow.Find(); err != nil {
		log.Fatal(err)
	}

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
