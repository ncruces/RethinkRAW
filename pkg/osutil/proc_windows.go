package osutil

import (
	"os"

	"golang.org/x/sys/windows"
)

func setPriority(proc os.Process, prio PriorityClass) error {
	const da = windows.PROCESS_SET_INFORMATION
	h, err := windows.OpenProcess(da, false, uint32(proc.Pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	var class uint32
	switch {
	case prio <= -20:
		class = windows.REALTIME_PRIORITY_CLASS
	case prio < -12:
		class = windows.HIGH_PRIORITY_CLASS
	case prio < -4:
		class = windows.ABOVE_NORMAL_PRIORITY_CLASS
	case prio < 4:
		class = windows.NORMAL_PRIORITY_CLASS
	case prio < 12:
		class = windows.BELOW_NORMAL_PRIORITY_CLASS
	default:
		class = windows.IDLE_PRIORITY_CLASS
	}
	return windows.SetPriorityClass(h, class)
}
