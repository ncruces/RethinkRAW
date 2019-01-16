package main

import (
	"log"
	"os"
	"path/filepath"

	pkg "github.com/ncruces/go-exiftool"
)

const exiftool = "./utils/exiftool"

var exifserver *pkg.Stayopen

func setupExifTool() *pkg.Stayopen {
	var err error
	os.Setenv("PAR_GLOBAL_TEMP", filepath.Join(dataDir, "exiftool"))
	exifserver, err = pkg.NewStayOpen(exiftool)
	if err != nil {
		log.Fatal(err)
	}
	return exifserver
}

func getMeta(path string) ([]byte, error) {
	log.Printf("exiftool [-ignoreMinorErrors -fixBase -groupHeadings0:1 %s]\n", path)
	return exifserver.ExtractFlags(path, "-ignoreMinorErrors", "-fixBase", "-groupHeadings0:1")
}

func copyMeta(orig, dest, name string) (err error) {
	opts := []string{"-tagsFromFile", orig, "-MakerNotes"}
	if name != "" {
		opts = append(opts, "-OriginalRawFileName-="+filepath.Base(orig), "-OriginalRawFileName="+filepath.Base(name))
	}
	opts = append(opts, "-overwrite_original", dest)

	log.Printf("exiftool %v\n", opts)
	_, err = exifserver.ExtractFlags("", opts...)
	return err
}
