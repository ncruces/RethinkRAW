//go:build !windows && !darwin

package dngconv

import (
	"bytes"
	"os"
	"os/exec"
)

func CheckInstall() error {
	out, err := exec.Command("wine", "cmd", "/c", "echo", "%ProgramW6432%").Output()
	if err != nil {
		return err
	}
	out = bytes.TrimRight(out, "\r\n")

	converter := string(out) + `\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
	out, err = exec.Command("winepath", converter).Output()
	if err != nil {
		return err
	}
	out = bytes.TrimRight(out, "\n")

	_, err = os.Stat(string(out))
	if err != nil {
		return err
	}

	conv = "wine"
	arg1 = converter
	return nil
}

var dngPathCache = map[string]string{}

func dngPath(path string) (string, error) {
	p, ok := dngPathCache[path]
	if ok {
		return p, nil
	}

	out, err := exec.Command("winepath", "-w", path).Output()
	if err != nil {
		return "", err
	}
	p = string(bytes.TrimRight(out, "\n"))

	if len(dngPathCache) > 100 {
		for k := range dngPathCache {
			delete(dngPathCache, k)
			break
		}
	}
	dngPathCache[path] = p
	return p, nil
}
