package craw

import (
	"testing"
)

func TestFujifilmCameraProfiles(t *testing.T) {
	// EmbedProfiles = "/Users/cruces/Desktop/Adobe DNG Converter.exe"
	profiles, err := FujifilmCameraProfiles("FinePix X100")
	if err != nil {
		t.Error(err)
	} else if len(profiles) != 7 {
		t.Errorf("Expected 7 profiles got %d", len(profiles))
	}
}
