package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

var (
	baseDir, dataDir, tempDir       string
	dcraw, exiftoolExe, exiftoolArg string
)

var serverConfig struct {
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
	dcraw = baseDir + "/utils/dcraw"
	exiftoolExe = baseDir + "/utils/exiftool/exiftool"
	if runtime.GOOS == "windows" {
		exiftoolArg = exiftoolExe
	}

	if serverConfig.DNGConverter == "" {
		switch runtime.GOOS {
		case "windows":
			serverConfig.DNGConverter = `C:\Program Files\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
		case "darwin":
			serverConfig.DNGConverter = "/Applications/Adobe DNG Converter.app/Contents/MacOS/Adobe DNG Converter"
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
