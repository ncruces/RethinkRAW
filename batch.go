package main

import (
	"compress/flate"
	"context"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/rethinkraw/pkg/osutil"
	"golang.org/x/sync/errgroup"
)

func toBatchPath(paths ...string) string {
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

func fromBatchPath(path string) []string {
	b64 := base64.NewDecoder(base64.RawURLEncoding, strings.NewReader(strings.TrimPrefix(path, "/")))
	flt := flate.NewReader(b64)
	var buf strings.Builder
	_, err := io.Copy(&buf, flt)
	if err != nil {
		return nil
	}
	str := buf.String()
	return strings.Split(strings.TrimSuffix(str, "\x00"), "\x00")
}

type batchPhoto struct {
	Path string
	Name string
}

func findPhotos(batch []string) ([]batchPhoto, error) {
	var photos []batchPhoto
	for _, path := range batch {
		var prefix string
		if len(batch) > 1 {
			prefix, _ = filepath.Split(path)
		} else {
			prefix = path + string(filepath.Separator)
		}
		err := filepath.WalkDir(path, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if osutil.HiddenFile(entry) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.Type().IsRegular() {
				if _, ok := extensions[strings.ToUpper(filepath.Ext(path))]; ok {
					var name string
					if strings.HasPrefix(path, prefix) {
						name = path[len(prefix):]
					} else {
						_, name = filepath.Split(path)
					}
					photos = append(photos, batchPhoto{path, name})
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

func batchProcess(ctx context.Context, photos []batchPhoto, proc func(ctx context.Context, photo batchPhoto) error) <-chan error {
	const parallelism = 6
	output := make(chan error, parallelism)

	go func() {
		group, ctx := errgroup.WithContext(ctx)
		group.SetLimit(parallelism)
		for _, photo := range photos {
			photo := photo
			group.Go(func() error {
				output <- proc(ctx, photo)
				return nil
			})
		}
		group.Wait()
		close(output)
	}()

	return output
}
