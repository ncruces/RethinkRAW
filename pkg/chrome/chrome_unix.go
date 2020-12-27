// +build !windows
// +build !darwin

package chrome

import (
	"os"
	"os/exec"
)

func findChrome() {
	versions := []string{"google-chrome-stable", "google-chrome", "chromium-browser", "chromium"}

	for _, v := range versions {
		if c, err := exec.LookPath(v); err == nil {
			chrome = c
			return
		}
	}
}

func exitProcess(p *os.Process) error {
	return p.Signal(os.Interrupt)
}
