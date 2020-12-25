// Package craw provides support for loading Camera Raw settings.
package craw

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Paths used to find Camera Raw settings.
// The defaults are fine if you have a recent version of DNG Converter installed.
var (
	GlobalSettings string // The global Camera Raw settings directory.
	UserSettings   string // The user's Camera Raw settings directory.
	EmbedProfiles  string // The file where to look for embed profiles (DNG Converter by default).
)

const (
	globalPrefixWin = "C:/ProgramData/Adobe/CameraRaw/"
	globalPrefixMac = "/Library/Application Support/Adobe/CameraRaw/"
)

func init() {
	switch runtime.GOOS {
	case "windows":
		GlobalSettings = os.Getenv("PROGRAMDATA") + `\Adobe\CameraRaw`
		UserSettings = os.Getenv("APPDATA") + `\Adobe\CameraRaw`
		EmbedProfiles = os.Getenv("PROGRAMFILES") + `\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
	case "darwin":
		GlobalSettings = "/Library/Application Support/Adobe/CameraRaw"
		UserSettings = os.Getenv("HOME") + "/Library/Application Support/Adobe/CameraRaw"
		EmbedProfiles = "/Applications/Adobe DNG Converter.app/Contents/MacOS/Adobe DNG Converter"
	}
}

func fixPath(path string) string {
	if strings.HasPrefix(path, globalPrefixWin) {
		path = filepath.Join(GlobalSettings, path[len(globalPrefixWin):])
	}
	if runtime.GOOS == "windows" && strings.HasPrefix(path, globalPrefixMac) {
		path = filepath.Join(GlobalSettings, path[len(globalPrefixMac):])
	}
	return filepath.FromSlash(path)
}
