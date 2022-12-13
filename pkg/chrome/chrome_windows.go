package chrome

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func findChrome() {
	versions := []string{`Google\Chrome`, `Chromium`, `Microsoft\Edge`}
	suffixes := []string{`Application\chrome.exe`, `Application\msedge.exe`}
	prefixes := []string{
		os.Getenv("LOCALAPPDATA"),
		os.Getenv("PROGRAMW6432"),
		os.Getenv("PROGRAMFILES"),
		os.Getenv("PROGRAMFILES(X86)"),
	}

	for _, v := range versions {
		for _, s := range suffixes {
			for _, p := range prefixes {
				if p != "" {
					c := filepath.Join(p, v, s)
					if _, err := os.Stat(c); err == nil {
						chrome = c
						return
					}
				}
			}
		}
	}
}

func signal(p *os.Process, sig os.Signal) error {
	if sig == syscall.SIGINT || sig == syscall.SIGHUP || sig == syscall.SIGTERM {
		return exec.Command("taskkill", "/pid", strconv.Itoa(p.Pid)).Run()
	}
	return p.Signal(sig)
}
