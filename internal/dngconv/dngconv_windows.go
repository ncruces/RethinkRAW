package dngconv

import (
	"context"
	"os"
	"os/exec"

	"github.com/ncruces/rethinkraw/pkg/osutil"
)

func findConverter() {
	const converter = `\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
	paths := []string{
		os.Getenv("PROGRAMW6432"),
		os.Getenv("PROGRAMFILES"),
		os.Getenv("PROGRAMFILES(X86)"),
	}
	for _, path := range paths {
		if path != "" {
			c := path + converter
			_, err := os.Stat(c)
			if err == nil {
				Path = c
				return
			}
		}
	}
}

func runConverter(ctx context.Context, args ...string) error {
	_, err := exec.CommandContext(ctx, Path, args...).Output()
	return err
}

func dngPath(path string) (string, error) {
	return osutil.GetANSIPath(path)
}
