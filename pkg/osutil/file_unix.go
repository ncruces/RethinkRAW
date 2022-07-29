//go:build !windows && !darwin

package osutil

import (
	"os"
	"os/exec"
)

func isHidden(os.DirEntry) bool {
	return false
}

func open(file string) error {
	return exec.Command("xdg-open", file).Run()
}

func getANSIPath(path string) (string, error) {
	return path, nil
}
