//go:build !windows && !darwin

package chrome

import (
	"os"
	"os/exec"
)

func findChrome() {
	versions := []string{"google-chrome-stable", "google-chrome", "chromium-browser", "chromium", "microsoft-edge-stable", "microsoft-edge"}

	for _, v := range versions {
		if c, err := exec.LookPath(v); err == nil {
			chrome = c
			return
		}
	}
}

func signal(p *os.Process, sig os.Signal) error {
	return p.Signal(sig)
}
