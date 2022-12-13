package dngconv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	conv string
	arg1 string
)

func Run(ctx context.Context, input, output string, args ...string) error {
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

	if arg1 != "" {
		args = append([]string{arg1}, args...)
	}
	args = append(args, "-d", dir, "-o", filepath.Base(output), input)

	cmd := exec.CommandContext(ctx, conv, args...)
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("DNG Converter: %w", err)
	}
	return nil
}
