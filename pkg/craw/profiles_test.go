package craw

import (
	"os"
	"runtime"
	"testing"
)

func init() {
	switch runtime.GOOS {
	case "windows":
		EmbedProfiles = os.Getenv("PROGRAMFILES") + `\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
	case "darwin":
		EmbedProfiles = "/Applications/Adobe DNG Converter.app/Contents/MacOS/Adobe DNG Converter"
	}
}

func TestGetCameraProfiles(t *testing.T) {
	profiles, err := GetCameraProfileNames("SONY", "ILCE-7")
	if err != nil {
		t.Error(err)
	} else if len(profiles) != 9 {
		t.Errorf("Expected 9 profiles got %d", len(profiles))
	}

	profiles, err = GetCameraProfileNames("FUJIFILM", "FinePix X100")
	if err != nil {
		t.Error(err)
	} else if len(profiles) != 8 {
		t.Errorf("Expected 8 profiles got %d", len(profiles))
	}
}
