package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

const dngconv = `C:\Program Files\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`

func toDng(input, output string, exp *exportSettings) error {
	err := os.RemoveAll(output)
	if err != nil {
		return err
	}

	dir := filepath.Dir(output)
	output = filepath.Base(output)

	opts := []string{}
	switch {
	case exp == nil:
		opts = append(opts, "-p2", "-side", "1920")
	case exp.Dng:
		if exp.Preview != "" {
			opts = append(opts, "-"+exp.Preview)
		}
		if exp.FastLoad {
			opts = append(opts, "-fl")
		}
		if exp.Embed {
			opts = append(opts, "-e")
		}
		if exp.Lossy {
			opts = append(opts, "-lossy")
		}
	default:
		opts = append(opts, "-p2")
	}
	opts = append(opts, "-d", dir, "-o", output, input)

	log.Printf("dngconv %v\n", opts)
	cmd := exec.Command(dngconv, opts...)
	if _, err := cmd.Output(); err != nil {
		return errors.WithMessagef(err, "DNG Converter: %v", opts)
	}
	return nil
}
