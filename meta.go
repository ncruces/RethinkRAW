package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	exif "./exiftool"
)

var exifserver *exif.Server

func setupExifTool() *exif.Server {
	var err error
	os.Setenv("PAR_GLOBAL_TEMP", filepath.Join(dataDir, "exiftool"))
	exifserver, err = exif.NewServer(exiftool)
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

func fixMetaJPEGAsync(orig string) (io.WriteCloser, *exif.AsyncResult) {
	opts := []string{"-tagsFromFile", orig, "-gps:all", "-exifIFD:all", "-commonIFD0", "-fast", "-"}

	rp, wp := io.Pipe()
	log.Printf("exiftool %v", opts)
	return wp, exif.CommandAsync(exiftool, rp, opts...)
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
