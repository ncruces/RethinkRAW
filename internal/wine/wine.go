package wine

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"sync"
)

var server struct {
	mtx sync.Mutex
	cmd *exec.Cmd
	err error
}

func Startup() error {
	server.mtx.Lock()
	defer server.mtx.Unlock()

	if server.cmd != nil {
		return server.err
	}

	server.cmd = exec.Command("wineserver", "--persistent", "--foreground", "--debug=0")
	server.err = server.cmd.Start()
	if server.err != nil {
		return server.err
	}

	server.err = exec.Command("wine", "cmd", "/c", "ver").Run()
	if server.err != nil {
		return server.err
	}

	go func() {
		err := server.cmd.Wait()

		server.mtx.Lock()
		defer server.mtx.Unlock()

		var eerr *exec.ExitError
		if err == nil || errors.As(err, &eerr) && eerr.ExitCode() == 2 {
			server.cmd = nil
		} else {
			server.err = err
		}
	}()

	return nil
}

func Shutdown() error {
	server.mtx.Lock()
	defer server.mtx.Unlock()

	if server.cmd == nil {
		return nil
	}
	return server.cmd.Process.Signal(os.Interrupt)
}

func Getenv(key string) (string, error) {
	out, err := exec.Command("wine", "cmd", "/c", "echo", "%"+key+"%").Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSuffix(out, []byte("\r\n"))), nil
}

func FromWindows(path string) (string, error) {
	out, err := exec.Command("winepath", "--unix", "-0", path).Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSuffix(out, []byte{0})), nil
}

func ToWindows(path string) (string, error) {
	out, err := exec.Command("winepath", "--windows", "-0", path).Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSuffix(out, []byte{0})), nil
}
