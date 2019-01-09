package main

import (
	"log"
	"os/exec"
	"path/filepath"
	"syscall"
)

const exiftool = "./utils/exiftool"

func getMeta(path string) ([]byte, error) {
	cmd := exec.Command(exiftool, "-ignoreMinorErrors", "-fixBase", "-groupHeadings0:1", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Output()
}

func fixMeta(path, dest, name string) (err error) {
	opts := []string{"-tagsFromFile", path, "-MakerNotes"}
	if name != "" {
		opts = append(opts, "-OriginalRawFileName-=orig.raw", "-OriginalRawFileName="+filepath.Base(name))
	}
	opts = append(opts, "-overwrite_original", dest)

	log.Printf("exiftool %v\n", opts)
	cmd := exec.Command(exiftool, opts...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err = cmd.Output()
	return err
}
