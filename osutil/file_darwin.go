package osutil

import (
	"os"
	"os/exec"
	"syscall"
)

func isHidden(fi os.FileInfo) bool {
	s, ok := fi.Sys().(*syscall.Stat_t)
	return ok && s.Flags&0x8000 != 0 // UF_HIDDEN
}

func open(file string) error {
	return exec.Command("open", file).Run()
}
