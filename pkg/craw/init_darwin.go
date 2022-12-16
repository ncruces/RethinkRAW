package craw

import "os"

func initPaths() {
	const path = "/Library/Application Support/Adobe/CameraRaw"
	if GlobalSettings == "" {
		GlobalSettings = path
	}
	if p, ok := os.LookupEnv("HOME"); UserSettings == "" && ok {
		UserSettings = p + path
	}
}
