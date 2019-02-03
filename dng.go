package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

func testDNGConverter(conv ...string) (err error) {
	for _, c := range append(conv, serverConfig.DNGConverter) {
		_, err = os.Stat(c)
		if err == nil {
			break
		}
	}
	return err
}

func runDNGConverter(input, output string, exp *exportSettings) error {
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
	case exp.DNG:
		if exp.Preview != "" {
			opts = append(opts, "-"+exp.Preview)
		}
		if exp.Lossy {
			opts = append(opts, "-lossy")
		}
		if exp.Embed {
			opts = append(opts, "-e")
		}
	default:
		opts = append(opts, "-p2")
	}
	opts = append(opts, "-d", dir, "-o", output, input)

	log.Printf("dngconv %v", opts)
	cmd := exec.Command(serverConfig.DNGConverter, opts...)
	if _, err := cmd.Output(); err != nil {
		return errors.WithMessagef(err, "DNG Converter")
	}
	return nil
}
