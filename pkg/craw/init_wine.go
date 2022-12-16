//go:build !windows && !darwin

package craw

import "github.com/ncruces/rethinkraw/pkg/wine"

func initPaths() {
	const path = `\Adobe\CameraRaw`
	if p, err := wine.Getenv("PROGRAMDATA"); GlobalSettings == "" && err == nil {
		GlobalSettings, _ = wine.FromWindows(p + path)
	}
	if p, err := wine.Getenv("APPDATA"); UserSettings == "" && err == nil {
		UserSettings, _ = wine.FromWindows(p + path)
	}
}
