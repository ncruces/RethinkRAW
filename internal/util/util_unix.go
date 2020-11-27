// +build !windows
// +build !darwin

package util

import (
	"os"
	"os/exec"
)

func FindChrome() string {
	versions := []string{"google-chrome-stable", "google-chrome", "chromium-browser", "chromium"}

	for _, v := range versions {
		if c, err := exec.LookPath(v); err == nil {
			return c
		}
	}
	return ""
}

func ExitChrome(cmd *exec.Cmd) {
	cmd.Process.Signal(os.Interrupt)
}

func GetANSIPath(path string) (string, error) {
	return path, nil
}

func BringToTop() {}

func HideConsole() {}
