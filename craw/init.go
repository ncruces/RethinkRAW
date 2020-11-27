// Package craw provides support for loading Camera Raw settings.
package craw

import (
	"os"
	"runtime"
)

var (
	GlobalSettings string
	UserSettings   string
)

func init() {
	switch runtime.GOOS {
	case "windows":
		GlobalSettings = os.Getenv("PROGRAMDATA") + `\Adobe\CameraRaw`
		UserSettings = os.Getenv("APPDATA") + `\Adobe\CameraRaw`
	case "darwin":
		GlobalSettings = "/Library/Application Support/Adobe/CameraRaw"
		UserSettings = os.Getenv("HOME") + "/Library/Application Support/Adobe/CameraRaw"
	}
}
