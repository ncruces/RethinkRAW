package osutil

import (
	"os"
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

	if os.Stdin.Fd() == 0 {
		h, _ := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)
		os.Stdin = os.NewFile(uintptr(h), "/dev/stdin")
		syscall.CloseOnExec(h)
	}
	if os.Stdout.Fd() == 0 {
		h, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
		os.Stdin = os.NewFile(uintptr(h), "/dev/stdout")
		syscall.CloseOnExec(h)
	}
	if os.Stderr.Fd() == 0 {
		h, _ := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)
		os.Stderr = os.NewFile(uintptr(h), "/dev/stderr")
		syscall.CloseOnExec(h)
	}

	if hwnd, _, _ := getConsoleWindow.Call(); hwnd != 0 {
		setForegroundWindow.Call(hwnd)
	}
	return nil
}
