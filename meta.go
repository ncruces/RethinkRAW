package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"rethinkraw/exiftool"
)

var exifserver *exiftool.Server

func setupExifTool() (*exiftool.Server, error) {
	var err error
	exifserver, err = exiftool.NewServer(exiftoolExe, exiftoolArg)
	return exifserver, err
}

func getMetaHTML(path string) ([]byte, error) {
	log.Print("exiftool (get meta)...")
	return exifserver.Command("-htmlFormat", "-groupHeadings", "-long", "-fixBase", "-ignoreMinorErrors", path)
}

func fixMetaDNG(orig, dest, name string) (err error) {
	opts := []string{"-tagsFromFile", orig, "-fixBase", "-MakerNotes", "-OriginalRawFileName-=" + filepath.Base(orig)}
	if name != "" {
		opts = append(opts, "-OriginalRawFileName="+filepath.Base(name))
	}
	opts = append(opts, "-overwrite_original", dest)

	log.Print("exiftool (fix dng)..")
	_, err = exifserver.Command(opts...)
	return err
}

func fixMetaJPEGAsync(orig string) (io.WriteCloser, io.ReadCloser, error) {
	// https://exiftool.org/forum/index.php?topic=8378.msg43043#msg43043
	opts := []string{"-tagsFromFile", orig, "-fixBase", "-CommonIFD0", "-ExifIFD:all", "-GPS:all", "-fast", "-"}

	inr, inw := io.Pipe()
	log.Print("exiftool (fix jpeg)...")
	out, err := exiftool.CommandAsync(exiftoolExe, exiftoolArg, inr, opts...)
	if err != nil {
		return nil, nil, err
	}
	return inw, out, nil
}

func hasEdits(path string) bool {
	log.Print("exiftool (has edits?)...")
	out, err := exifserver.Command("-XMP-photoshop:all", path)
	return err == nil && len(out) > 0
}

func tiffOrientation(path string) int {
	log.Print("exiftool (get orientation)...")
	out, err := exifserver.Command("--printConv", "-short3", "-fast2", "-Orientation", path)
	if err != nil {
		return 0
	}

	var orientation int
	_, err = fmt.Fscanf(bytes.NewReader(out), "%d\n", &orientation)
	if err != nil {
		return 0
	}

	return orientation
}
