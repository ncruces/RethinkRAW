package craw

import (
	"path/filepath"
	"testing"
)

func TestLoadIndex(t *testing.T) {
	tests := []string{
		"CameraProfiles/Index.dat",
		"LensProfiles/Index.dat",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			path := filepath.Join(GlobalSettings, filepath.FromSlash(tt))
			idx, err := LoadIndex(path)
			if err != nil {
				t.Error(err)
			} else {
				t.Logf("Read %d records.", len(idx))
			}
		})
	}
}
