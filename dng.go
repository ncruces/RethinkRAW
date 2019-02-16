package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

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
	cmd := exec.Command(serverConfig.DNGConverter, opts...)
	if _, err := cmd.Output(); err != nil {
		return errors.WithMessagef(err, "DNG Converter")
	}
	return nil
}
