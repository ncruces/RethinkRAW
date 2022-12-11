package osutil

import "os"

type PriorityClass int

const (
	Realtime    PriorityClass = -20
	High        PriorityClass = -16
	AboveNormal PriorityClass = -8
	Normal      PriorityClass = 0
	BelowNormal PriorityClass = 8
	Idle        PriorityClass = 16
)

// SetPriority sets the scheduling priority of proc.
func SetPriority(proc os.Process, prio PriorityClass) error {
	return setPriority(proc, prio)
}
