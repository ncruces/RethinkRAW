// Package craw provides support for loading Camera Raw settings.
package craw

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Paths used to find Camera Raw settings.
var (
	GlobalSettings string // The global Camera Raw settings directory.
	UserSettings   string // The user's Camera Raw settings directory.
	EmbedProfiles  string // The file where to look for embed profiles.
)

const (
	globalPrefixWin = "C:/ProgramData/Adobe/CameraRaw/"
	globalPrefixMac = "/Library/Application Support/Adobe/CameraRaw/"
)

var once sync.Once

func fixPath(path string) string {
	if strings.HasPrefix(path, globalPrefixWin) {
		path = filepath.Join(GlobalSettings, path[len(globalPrefixWin):])
	}
	if runtime.GOOS != "darwin" && strings.HasPrefix(path, globalPrefixMac) {
		path = filepath.Join(GlobalSettings, path[len(globalPrefixMac):])
	}
	return filepath.FromSlash(path)
}
