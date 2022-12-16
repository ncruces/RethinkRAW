package dngconv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var Path string

var once sync.Once

// IsInstalled checks if Adobe DNG Converter is installed.
// If true, [Path] will be set to the converter's executable path.
func IsInstalled() bool {
	once.Do(findConverter)
	return Path != ""
}

// Convert converts an input RAW file into an output DNG using Adobe DNG Converter.
func Convert(ctx context.Context, input, output string, args ...string) error {
	once.Do(findConverter)

	input, err := dngPath(input)
	if err != nil {
		return err
	}

	dir, err := dngPath(filepath.Dir(output))
	if err != nil {
		return err
	}

	err = os.RemoveAll(output)
	if err != nil {
		return err
	}

	args = append(args, "-d", dir, "-o", filepath.Base(output), input)
	err = runConverter(ctx, args...)
	if err != nil {
		return fmt.Errorf("dng converter: %w", err)
	}
	return nil
}
