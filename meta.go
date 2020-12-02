package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/ncruces/go-exiftool"
)

var exifserver *exiftool.Server

func setupExifTool() (s *exiftool.Server, err error) {
	exifserver, err = exiftool.NewServer()
	return exifserver, err
}

func getMetaHTML(path string) ([]byte, error) {
	log.Print("exiftool (get meta)...")
	return exifserver.Command("-htmlFormat", "-groupHeadings", "-long", "-fixBase", "-ignoreMinorErrors", path)
}

func fixMetaDNG(orig, dest, name string) error {
	opts := []string{"-tagsFromFile", orig, "-fixBase", "-MakerNotes", "-OriginalRawFileName-=" + filepath.Base(orig)}
	if name != "" {
		opts = append(opts, "-OriginalRawFileName="+filepath.Base(name))
	}
	opts = append(opts, "-overwrite_original", dest)

	log.Print("exiftool (fix dng)...")
	_, err := exifserver.Command(opts...)
	return err
}

func fixMetaJPEGAsync(orig string) (io.WriteCloser, io.ReadCloser, error) {
	// https://exiftool.org/forum/index.php?topic=8378.msg43043#msg43043
	opts := []string{"-tagsFromFile", orig, "-fixBase", "-CommonIFD0", "-ExifIFD:all", "-GPS:all", "-fast", "-"}

	log.Print("exiftool (fix jpeg)...")
	return exiftool.CommandAsync(opts...)
}

func hasEdits(path string) bool {
	log.Print("exiftool (has edits?)...")
	out, err := exifserver.Command("-XMP-photoshop:all", path)
	return err == nil && len(out) > 0
}

func rawOrientation(path string) int {
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

func cameraMatchingWhiteBalance(path string) string {
	log.Print("exiftool (get camera matching white balance)...")
	out, err := exifserver.Command("-duplicates", "-short3", "-fast", "-WhiteBalance", path)
	if err != nil {
		return ""
	}

	for scan := bufio.NewScanner(bytes.NewReader(out)); scan.Scan(); {
		switch wb := scan.Text(); wb {
		case "Auto", "Daylight", "Cloudy", "Shade", "Tungsten", "Fluorescent", "Flash":
			return wb
		case "Sunny":
			return "Daylight"
		case "Overcast":
			return "Cloudy"
		case "Incandescent":
			return "Tungsten"
		}
	}
	return "As Shot"
}
