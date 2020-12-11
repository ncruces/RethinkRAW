// +build !windows
// +build !darwin

package util

import (
	"os"
	"os/exec"
)

func GetANSIPath(path string) (string, error) {
	return path, nil
}

func BringToTop() {}

func HideConsole() {}
