package main

import (
	"os"
	"path/filepath"
	"runtime"
)

var (
	baseDir, dataDir, tempDir string
	exiftoolExe, exiftoolArg  string
	dcraw, dngConverter       string
	cameraRawPaths            []string
)

func setupPaths() (err error) {
	if exe, err := os.Executable(); err != nil {
		return err
	} else {
		baseDir = filepath.Dir(exe)
	}

	dataDir = filepath.Join(baseDir, "data")
	tempDir = filepath.Join(os.TempDir(), "RethinkRAW")

	tempDir, err = getANSIPath(tempDir)
	if err != nil {
		return err
	}

	dcraw = baseDir + "/utils/dcraw"
	exiftoolExe = baseDir + "/utils/exiftool/exiftool"
	if runtime.GOOS == "windows" {
		exiftoolArg = exiftoolExe
	}
	switch runtime.GOOS {
	case "windows":
		dngConverter = os.Getenv("PROGRAMFILES") + `\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
		cameraRawPaths = []string{
			os.Getenv("PROGRAMDATA") + `\Adobe\CameraRaw`,
			os.Getenv("APPDATA") + `\Adobe\CameraRaw`,
		}
	case "darwin":
		dngConverter = "/Applications/Adobe DNG Converter.app/Contents/MacOS/Adobe DNG Converter"
		cameraRawPaths = []string{
			"/Library/Application Support/Adobe/CameraRaw",
			os.Getenv("HOME") + "/Library/Application Support/Adobe/CameraRaw",
		}
	}

	testDir := func() error {
		if err := os.MkdirAll(dataDir, 0700); err != nil {
			return err
		}
		if f, err := os.Create(filepath.Join(dataDir, "lastrun")); err != nil {
			return err
		} else {
			return f.Close()
		}
	}
	if testDir() == nil {
		return nil
	}
	if data, err := os.UserConfigDir(); err != nil {
		return err
	} else {
		dataDir = filepath.Join(data, "RethinkRAW")
	}
	return testDir()
}
