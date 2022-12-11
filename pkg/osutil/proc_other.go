//go:build !windows

package osutil

import (
	"os"
	"syscall"
)

func setPriority(proc os.Process, prio PriorityClass) error {
	return syscall.Setpriority(syscall.PRIO_PROCESS, proc.Pid, int(prio))
}
