package craw

import "os"

func initPaths() {
	const path = `\Adobe\CameraRaw`
	if p, ok := os.LookupEnv("PROGRAMDATA"); GlobalSettings == "" && ok {
		GlobalSettings = p + path
	}
	if p, ok := os.LookupEnv("APPDATA"); UserSettings == "" && ok {
		UserSettings = p + path
	}
}
