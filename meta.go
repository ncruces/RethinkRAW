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

func getMeta(path string) ([]byte, error) {
	log.Print("exiftool (get meta)...")
	return exifserver.Command("-ignoreMinorErrors", "-fixBase", "-htmlFormat", "-groupHeadings", "-long", path)
	//return exifserver.Command("-ignoreMinorErrors", "-fixBase", "-json", "-groupHeadings", "-long", path)
}

func fixMetaDNG(orig, dest, name string) (err error) {
	opts := []string{"-tagsFromFile", orig, "-MakerNotes", "-OriginalRawFileName-=" + filepath.Base(orig)}
	if name != "" {
		opts = append(opts, "-OriginalRawFileName="+filepath.Base(name))
	}
	opts = append(opts, "-overwrite_original", dest)

	log.Print("exiftool (fix dng)..")
	_, err = exifserver.Command(opts...)
	return err
}

func fixMetaJPEGAsync(orig string) (io.WriteCloser, *exiftool.AsyncResult) {
	opts := []string{"-tagsFromFile", orig, "-GPS:all", "-ExifIFD:all", "-CommonIFD0", "-fast", "-"}

	rp, wp := io.Pipe()
	log.Print("exiftool (fix jpeg)...")
	return wp, exiftool.CommandAsync(exiftoolExe, exiftoolArg, rp, opts...)
}

func hasEdits(path string) bool {
	log.Print("exiftool (has edits?)...")
	out, err := exifserver.Command("-XMP-photoshop:*", path)
	return err == nil && len(out) > 0
}

func tiffOrientation(path string) int {
	log.Print("exiftool (get orientation)...")
	out, err := exifserver.Command("-s3", "-n", "-Orientation", "-fast2", path)
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
