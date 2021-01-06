package osutil

import (
	"os/exec"
	"syscall"
)

func createConsole() error {
	if hwnd, _, _ := getConsoleWindow.Call(); hwnd != 0 {
		return nil
	}

	cmd := exec.Command("cmd", "/c", "pause")
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		return err
	}

	var buf [32]byte
	_, err = out.Read(buf[:])
	if err == nil {
		s, _, gerr := attachConsole.Call(uintptr(cmd.Process.Pid))
		if s == 0 {
			err = gerr
		}
	}
	if cerr := in.Close(); err == nil {
		err = cerr
	}
	if werr := cmd.Wait(); err == nil {
		err = werr
	}
	if err != nil {
		return err
	}

	if hwnd, _, _ := getConsoleWindow.Call(); hwnd != 0 {
		setForegroundWindow.Call(hwnd)
	}
	return nil
}
