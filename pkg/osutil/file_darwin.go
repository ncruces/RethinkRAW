package osutil

import (
	"os"
	"os/exec"
	"syscall"
)

func isHidden(de os.DirEntry) bool {
	i, err := de.Info()
	if err != nil {
		return false
	}
	s, ok := i.Sys().(*syscall.Stat_t)
	return ok && s.Flags&0x8000 != 0 // UF_HIDDEN
}

func open(file string) error {
	return exec.Command("open", file).Run()
}

func getANSIPath(path string) (string, error) {
	return path, nil
}
