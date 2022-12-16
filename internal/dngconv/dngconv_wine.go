//go:build !windows && !darwin

package dngconv

import (
	"context"
	"os"

	"github.com/ncruces/rethinkraw/pkg/wine"
)

func findConverter() {
	const converter = `\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
	paths := []string{
		"PROGRAMW6432",
		"PROGRAMFILES",
		"PROGRAMFILES(X86)",
	}
	for _, path := range paths {
		env, err := wine.Getenv(path)
		if err != nil {
			continue
		}

		unix, err := wine.FromWindows(env + converter)
		if err != nil {
			continue
		}

		_, err = os.Stat(unix)
		if err == nil {
			Path = unix
			break
		}
	}
}

func runConverter(ctx context.Context, args ...string) error {
	_, err := wine.CommandContext(ctx, Path, args...).Output()
	return err
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
