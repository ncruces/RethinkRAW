package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

const dcraw = "./utils/dcraw"
const jpegtran = "./utils/jpegtran"
const exiftoolExe = "./utils/exiftool/exiftool"
const exiftoolArg = "./utils/exiftool/exiftool"

var baseDir, dataDir, tempDir string
var serverConfig serverSettings

type serverSettings struct {
	DNGConverter string `json:"dngConverter"`
}

func loadConfig() error {
	if f, err := os.Open(filepath.Join(dataDir, "config.json")); os.IsNotExist(err) {
		// use defaults
	} else if err != nil {
		return err
	} else {
		dec := json.NewDecoder(f)
		if err := dec.Decode(&serverConfig); err != nil {
			return err
		}
	}

	// set defaults
	if serverConfig.DNGConverter == "" {
		switch runtime.GOOS {
		case "windows":
			serverConfig.DNGConverter = `C:\Program Files\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
		case "darwin":
			serverConfig.DNGConverter = `/Applications/Adobe DNG Converter.app/Contents/MacOS/Adobe DNG Converter`
		}
	}
	return nil
}

func saveConfig() error {
	if f, err := os.Create(filepath.Join(dataDir, "config.json")); err != nil {
		return err
	} else {
		enc := json.NewEncoder(f)
		return enc.Encode(&serverConfig)
	}
}
