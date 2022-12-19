// Package wine provides support to run Windows programs under Wine.
package wine

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
)

var server struct {
	wg  sync.WaitGroup
	mtx sync.Mutex
	cmd *exec.Cmd
}

// IsInstalled checks if Wine is installed.
func IsInstalled() bool {
	_, err := exec.LookPath("wine")
	return err == nil
}

// Startup starts a persistent Wine server.
// A persistent Wine server improves the performance and reliability of subsequent usages of Wine.
func Startup() error {
	server.mtx.Lock()
	defer server.mtx.Unlock()

	server.wg.Add(1)
	server.cmd = exec.Command("wineserver", "--persistent", "--foreground", "--debug=0")
	if err := server.cmd.Start(); err != nil {
		return err
	}

	go func() {
		exec.Command("wine", "cmd", "/c", "ver").Run()
		server.wg.Done()

		err := server.cmd.Wait()

		server.mtx.Lock()
		defer server.mtx.Unlock()

		var eerr *exec.ExitError
		if err == nil || errors.As(err, &eerr) && eerr.ExitCode() == 2 {
			server.cmd = nil
		}
	}()

	return nil
}

// Shutdown shuts any persistent Wine server started by this package down.
func Shutdown() error {
	server.mtx.Lock()
	defer server.mtx.Unlock()

	cmd := server.cmd
	if cmd != nil {
		server.cmd = nil
		return server.cmd.Process.Signal(os.Interrupt)
	}
	return nil
}

// Getenv retrieves the value of the Windows (Wine) environment variable named by the key.
func Getenv(key string) (string, error) {
	server.wg.Wait()
	out, err := exec.Command("wine", "cmd", "/c", "echo", "%"+key+"%").Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSuffix(out, []byte("\r\n"))), nil
}

// FromWindows translates a Windows (Wine) to a Unix path.
func FromWindows(path string) (string, error) {
	server.wg.Wait()
	out, err := exec.Command("winepath", "--unix", "-0", path).Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSuffix(out, []byte{0})), nil
}

// ToWindows translates a Unix to a Windows (Wine) path.
func ToWindows(path string) (string, error) {
	server.wg.Wait()
	out, err := exec.Command("winepath", "--windows", "-0", path).Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSuffix(out, []byte{0})), nil
}

// Command returns the [exec.Cmd] struct to execute a Windows program using Wine.
func Command(name string, args ...string) *exec.Cmd {
	args = append([]string{name}, args...)
	return exec.Command("wine", args...)
}

// CommandContext is like [Command] but includes a context.
func CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	args = append([]string{name}, args...)
	return exec.CommandContext(ctx, "wine", args...)
}
