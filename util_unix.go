// +build !windows

package main

import (
	"os"
	"os/exec"
	"strings"
)

func findChrome() string {
	return ""
}

func exitChrome(cmd *exec.Cmd) {}

func openURLCmd(url string) *exec.Cmd {
	return exec.Command("xdg-open", url)
}

func isHidden(fi os.FileInfo) bool {
	return strings.HasPrefix(fi.Name(), ".")
}

func getANSIPath(path string) (string, error) {
	return path, nil
}

func hideConsole() {}

func bringToTop() {}

func handleConsoleCtrl(c chan<- os.Signal) {}
