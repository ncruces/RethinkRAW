package main

import (
	"log"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
)

const dngconv = `C:\Program Files\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`

func toDng(input, output, dir string, preview, lossy bool) error {
	opts := []string{}
	switch {
	case preview:
		opts = append(opts, "-side", "1920", "-lossy")
	case lossy:
		opts = append(opts, "-lossy")
	}
	opts = append(opts, "-p2", "-fl", "-d", dir, "-o", output, input)

	log.Printf("dngconv %v\n", opts)
	cmd := exec.Command(dngconv, opts...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if _, err := cmd.Output(); err != nil {
		return errors.WithMessagef(err, "DNG Converter: %v", opts)
	}
	return nil
}
