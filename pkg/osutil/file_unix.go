// +build !windows
// +build !darwin

package osutil

import (
	"os"
	"os/exec"
)

func isHidden(fi os.FileInfo) bool {
	return false
}

func open(file string) error {
	return exec.Command("xdg-open", file).Run()
}
