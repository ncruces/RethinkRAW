package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const dcraw = "./utils/dcraw"
const exiv2 = "./utils/exiv2"
const exiftool = "./utils/exiftool"
const jpegtran = "./utils/jpegtran"

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
		serverConfig.DNGConverter = `C:\Program Files\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
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
