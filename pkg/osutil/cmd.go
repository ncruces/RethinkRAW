package osutil

import (
	"os"
	"runtime"
	"strings"
)

// CreateConsole ensures a Windows process has an attached console.
// If needed, it creates an hidden console and attaches to it.
func CreateConsole() error {
	return createConsole()
}

// CleanupArgs cleans up os.Args.
// On macOS, it removes the Process Serial Number arg.
func CleanupArgs() {
	if runtime.GOOS == "darwin" {
		for i, a := range os.Args {
			if strings.HasPrefix(a, "-psn_") {
				os.Args = append(os.Args[:i], os.Args[i+1:]...)
				break
			}
		}
	}
}
