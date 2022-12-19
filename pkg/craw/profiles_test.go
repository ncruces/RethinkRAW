package craw

import (
	"testing"

	"github.com/ncruces/rethinkraw/pkg/dngconv"
)

func TestGetCameraProfiles(t *testing.T) {
	profiles, err := GetCameraProfileNames("SONY", "ILCE-7")
	if err != nil {
		t.Error(err)
	} else if len(profiles) != 9 {
		t.Errorf("Expected 9 profiles got %d", len(profiles))
	}

	if dngconv.IsInstalled() {
		EmbedProfiles = dngconv.Path
		profiles, err = GetCameraProfileNames("FUJIFILM", "FinePix X100")
		if err != nil {
			t.Error(err)
		} else if len(profiles) != 8 {
			t.Errorf("Expected 8 profiles got %d", len(profiles))
		}
	}
}
