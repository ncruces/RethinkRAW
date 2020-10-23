package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func testDNGConverter() error {
	_, err := os.Stat(dngConverter)
	return err
}

func runDNGConverter(input, output string, side int, exp *exportSettings) error {
	err := os.RemoveAll(output)
	if err != nil {
		return err
	}

	dir := filepath.Dir(output)
	output = filepath.Base(output)

	opts := []string{}
	if exp != nil && exp.DNG {
		if exp.Preview != "" {
			opts = append(opts, "-"+exp.Preview)
		}
		if exp.Lossy {
			opts = append(opts, "-lossy")
		}
		if exp.Embed {
			opts = append(opts, "-e")
		}
	} else {
		if side > 0 {
			opts = append(opts, "-lossy", "-side", strconv.Itoa(side))
		}
		opts = append(opts, "-p2")
	}
	opts = append(opts, "-d", dir, "-o", output, input)

	log.Print("dng converter...")
	cmd := exec.Command(dngConverter, opts...)
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("DNG Converter: %w", err)
	}
	return nil
}
