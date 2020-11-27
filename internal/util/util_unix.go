// +build !windows
// +build !darwin

package main

import (
	"os"
	"os/exec"
)

func findChrome() string {
	versions := []string{"google-chrome-stable", "google-chrome", "chromium-browser", "chromium"}

	for _, v := range versions {
		if c, err := exec.LookPath(v); err == nil {
			return c
		}
	}
	return ""
}

func exitChrome(cmd *exec.Cmd) {
	cmd.Process.Signal(os.Interrupt)
}

func getANSIPath(path string) (string, error) {
	return path, nil
}

func bringToTop() {}

func hideConsole() {}
