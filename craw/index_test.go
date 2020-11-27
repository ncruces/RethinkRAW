// Package craw provides support for loading Camera Raw settings.
package craw

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadIndex(t *testing.T) {
	tests := []string{
		"CameraProfiles/Index.dat",
		"LensProfiles/Index.dat",
	}

	var craw string
	switch runtime.GOOS {
	case "windows":
		craw = os.Getenv("PROGRAMDATA") + `\Adobe\CameraRaw`
	case "darwin":
		craw = "/Library/Application Support/Adobe/CameraRaw"
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			path := filepath.Join(craw, filepath.FromSlash(tt))
			idx, err := LoadIndex(path)
			if err != nil {
				t.Error(err)
			} else {
				t.Logf("Read %d records.", len(idx))
			}
		})
	}
}
