package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ncruces/go-exiftool"
)

var (
	baseDir, dataDir, tempDir string
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

	switch runtime.GOOS {
	case "windows":
		dcraw = baseDir + `\utils\dcraw.exe`
		exiftool.Exec = baseDir + `\utils\exiftool\exiftool.exe`
		exiftool.Arg1 = strings.TrimSuffix(exiftool.Exec, ".exe")
		exiftool.Config = baseDir + `\utils\ExifTool_config.pl`
		dngConverter = os.Getenv("PROGRAMFILES") + `\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
		cameraRawPaths = []string{
			os.Getenv("PROGRAMDATA") + `\Adobe\CameraRaw`,
			os.Getenv("APPDATA") + `\Adobe\CameraRaw`,
		}
	case "darwin":
		dcraw = baseDir + "/utils/dcraw"
		exiftool.Exec = baseDir + "/utils/exiftool/exiftool"
		exiftool.Config = baseDir + "/utils/ExifTool_config.pl"
		dngConverter = "/Applications/Adobe DNG Converter.app/Contents/MacOS/Adobe DNG Converter"
		cameraRawPaths = []string{
			"/Library/Application Support/Adobe/CameraRaw",
			os.Getenv("HOME") + "/Library/Application Support/Adobe/CameraRaw",
		}
	}

	if testDataDir() == nil {
		return nil
	}
	if data, err := os.UserConfigDir(); err != nil {
		return err
	} else {
		dataDir = filepath.Join(data, "RethinkRAW")
	}
	return testDataDir()
}

func testDataDir() error {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return err
	}
	if f, err := os.Create(filepath.Join(dataDir, "lastrun")); err != nil {
		return err
	} else {
		return f.Close()
	}
}
