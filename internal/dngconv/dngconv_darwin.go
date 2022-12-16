package dngconv

import (
	"context"
	"os"
	"os/exec"
)

func findConverter() {
	const converter = "/Applications/Adobe DNG Converter.app/Contents/MacOS/Adobe DNG Converter"
	_, err := os.Stat(converter)
	if err != nil {
		return
	}
	Path = converter
}

func runConverter(ctx context.Context, args ...string) error {
	_, err := exec.CommandContext(ctx, Path, args...).Output()
	return err
}

func dngPath(path string) (string, error) {
	return path, nil
}
