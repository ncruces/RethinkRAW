package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	// ignore Process Serial Number argument
	for i, a := range os.Args {
		if strings.HasPrefix(a, "-psn_") {
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			break
		}
	}
}

func FindChrome() string {
	versions := []string{"Google Chrome", "Chromium"}

	for _, v := range versions {
		c := filepath.Join("/Applications", v+".app", "Contents/MacOS", v)
		if _, err := os.Stat(c); err == nil {
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
