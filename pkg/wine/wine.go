// Package wine provides support to run Windows programs under Wine.
package wine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

var server struct {
	wg  sync.WaitGroup
	cmd *exec.Cmd
}

// IsInstalled checks if Wine is installed.
func IsInstalled() bool {
	_, err := exec.LookPath("wine")
	return err == nil
}

// Startup starts a persistent Wine server,
// which improves the performance and reliability of subsequent usages of Wine.
func Startup() error {
	cmd := exec.Command("wineserver", "--persistent", "--foreground", "--debug=0")
	if err := cmd.Start(); err != nil {
		return err
	}

	server.cmd = cmd
	server.wg.Add(1)
	go func() {
		// This ensures commands that capture stdout succeed.
		exec.Command("wine", "cmd", "/c", "ver").Run()
		server.wg.Done()
	}()

	return nil
}

// Shutdown shuts down a persistent Wine server started by this package,
// and waits for it to complete.
func Shutdown() (ex error) {
	if server.cmd != nil {
		server.cmd.Process.Signal(os.Interrupt)
		err := server.cmd.Wait()
		server.cmd = nil

		var eerr *exec.ExitError
		if errors.As(err, &eerr) && eerr.ExitCode() == 2 {
			return nil
		}
		return err
	}
	return nil
}

// Getenv retrieves the value of the Windows (Wine) environment variable named by the key.
func Getenv(key string) (string, error) {
	for _, b := range []byte(key) {
		if b == '(' || b == ')' || b == '_' ||
			'0' <= b && b <= '9' ||
			'a' <= b && b <= 'z' ||
			'A' <= b && b <= 'Z' {
			continue
		}
		return "", fmt.Errorf("wine: invalid character %q in variable name", b)
	}

	server.wg.Wait()
	out, err := exec.Command("wine", "cmd", "/c", "if defined", key, "echo", "%"+key+"%").Output()
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
	server.wg.Wait()
	args = append([]string{name}, args...)
	return exec.Command("wine", args...)
}

// CommandContext is like [Command] but includes a context.
func CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	server.wg.Wait()
	args = append([]string{name}, args...)
	return exec.CommandContext(ctx, "wine", args...)
}
