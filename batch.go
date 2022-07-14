package main

import (
	"compress/flate"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ncruces/rethinkraw/pkg/osutil"
)

func EncodeBatch(paths []string) string {
	var buf strings.Builder
	b64 := base64.NewEncoder(base64.RawURLEncoding, &buf)
	flt, err := flate.NewWriter(b64, flate.BestCompression)
	if err != nil {
		panic(err)
	}
	for _, path := range paths {
		flt.Write([]byte(path))
		flt.Write([]byte{0})
	}
	flt.Close()
	b64.Close()
	return buf.String()
}

func DecodeBatch(batch string) []string {
	b64 := base64.NewDecoder(base64.RawURLEncoding, strings.NewReader(batch))
	flt := flate.NewReader(b64)
	var buf strings.Builder
	_, err := io.Copy(&buf, flt)
	if err != nil {
		return nil
	}
	str := buf.String()
	return strings.Split(strings.TrimRight(str, "\x00"), "\x00")
}

type BatchPhoto struct {
	Path string
	Name string
}

func FindPhotos(batch []string) ([]BatchPhoto, error) {
	var photos []BatchPhoto
	for _, path := range batch {
		var prefix string
		if len(batch) > 1 {
			prefix, _ = filepath.Split(path)
		} else {
			prefix = path + string(filepath.Separator)
		}
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if osutil.HiddenFile(info) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.Mode().IsRegular() {
				if _, ok := extensions[strings.ToUpper(filepath.Ext(path))]; ok {
					var name string
					if strings.HasPrefix(path, prefix) {
						name = path[len(prefix):]
					} else {
						_, name = filepath.Split(path)
					}
					photos = append(photos, BatchPhoto{path, name})
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return photos, nil
}

func BatchProcess(photos []BatchPhoto, proc func(photo BatchPhoto) error) <-chan error {
	const parallelism = 2

	output := make(chan error, parallelism)
	input := make(chan BatchPhoto)
	wait := sync.WaitGroup{}
	wait.Add(parallelism)

	for n := 0; n < parallelism; n++ {
		go func() {
			for photo := range input {
				output <- proc(photo)
			}
			wait.Done()
		}()
	}

	go func() {
		for _, photo := range photos {
			input <- photo
		}
		close(input)
		wait.Wait()
		close(output)
	}()

	return output
}
