//go:build !windows && !darwin

package dngconv

import (
	"os"

	"github.com/ncruces/rethinkraw/internal/wine"
)

func CheckInstall() error {
	programs, err := wine.Getenv("ProgramW6432")
	if err != nil {
		return err
	}

	converter := programs + `\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`

	file, err := wine.FromWindows(converter)
	if err != nil {
		return err
	}

	_, err = os.Stat(file)
	if err != nil {
		return err
	}

	conv = "wine"
	arg1 = converter
	return nil
}

var dngPathCache = map[string]string{}

func dngPath(path string) (string, error) {
	p, ok := dngPathCache[path]
	if ok {
		return p, nil
	}

	p, err := wine.ToWindows(path)
	if err != nil {
		return "", err
	}

	if len(dngPathCache) > 100 {
		for k := range dngPathCache {
			delete(dngPathCache, k)
			break
		}
	}
	dngPathCache[path] = p
	return p, nil
}
