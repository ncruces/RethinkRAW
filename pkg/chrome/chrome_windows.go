package chrome

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func findChrome() {
	versions := []string{`Google\Chrome`, `Chromium`}
	prefixes := []string{os.Getenv("LOCALAPPDATA"), os.Getenv("PROGRAMFILES"), os.Getenv("PROGRAMFILES(X86)")}
	suffix := `\Application\chrome.exe`

	for _, v := range versions {
		for _, p := range prefixes {
			if p != "" {
				c := filepath.Join(p, v, suffix)
				if _, err := os.Stat(c); err == nil {
					chrome = c
					return
				}
			}
		}
	}
}

func signal(p *os.Process, sig os.Signal) error {
	if sig == syscall.SIGINT || sig == syscall.SIGTERM {
		pid := strconv.Itoa(p.Pid)
		err := exec.Command("taskkill", "/pid", pid).Run()
		if sig == syscall.SIGINT {
			return err
		}

		go func() {
			time.Sleep(time.Second)
			exec.Command("taskkill", "/f", "/t", "/pid", pid).Run()
		}()
		_, err = p.Wait()
		return err
	}
	return p.Signal(sig)
}
