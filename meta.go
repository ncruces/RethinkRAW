package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"./exiftool"
)

var exifserver *exiftool.Server

func setupExifTool() *exiftool.Server {
	var err error
	exifserver, err = exiftool.NewServer(exiftoolExe, exiftoolArg)
	if err != nil {
		log.Fatal(err)
	}
	return exifserver
}

func getMeta(path string) ([]byte, error) {
	log.Printf("exiftool [-ignoreMinorErrors -fixBase -groupHeadings0:1 %s]", path)
	return exifserver.Command("-ignoreMinorErrors", "-fixBase", "-groupHeadings0:1", path)
}

func fixMetaDNG(orig, dest, name string) (err error) {
	opts := []string{"-tagsFromFile", orig, "-makerNotes"}
	if name != "" {
		opts = append(opts, "-originalRawFileName-="+filepath.Base(orig), "-originalRawFileName="+filepath.Base(name))
	}
	opts = append(opts, "-overwrite_original", dest)

	log.Printf("exiftool %v", opts)
	_, err = exifserver.Command(opts...)
	return err
}

func fixMetaJPEGAsync(orig string) (io.WriteCloser, *exiftool.AsyncResult) {
	opts := []string{"-tagsFromFile", orig, "-gps:all", "-exifIFD:all", "-commonIFD0", "-fast", "-"}

	rp, wp := io.Pipe()
	log.Printf("exiftool %v", opts)
	return wp, exiftool.CommandAsync(exiftoolExe, exiftoolArg, rp, opts...)
}

func hasEdits(path string) bool {
	log.Printf("exiftool [-xmp-photoshop:* %s]", path)
	out, err := exifserver.Command("-xmp-photoshop:*", path)
	return err == nil && len(out) > 0
}

func tiffOrientation(path string) int {
	log.Printf("exiftool [-s3 -n -orientation %s]", path)
	out, err := exifserver.Command("-s3", "-n", "-orientation", path)
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
