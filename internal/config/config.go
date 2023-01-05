package config

import (
	"mime"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ncruces/go-exiftool"
	"github.com/ncruces/rethinkraw/pkg/dcraw"
)

var (
	ServerMode                bool
	BaseDir, DataDir, TempDir string
)

func init() {
	mime.AddExtensionType(".dng", "image/x-adobe-dng")
}

func SetupPaths() (err error) {
	if exe, err := os.Executable(); err != nil {
		return err
	} else if exe, err := filepath.EvalSymlinks(exe); err != nil {
		return err
	} else {
		BaseDir = filepath.Dir(exe)
	}

	DataDir = filepath.Join(BaseDir, "data")
	TempDir = filepath.Join(os.TempDir(), "RethinkRAW")

	name := filepath.Base(os.Args[0])
	if runtime.GOOS == "windows" {
		ServerMode = name == "RethinkRAW.com"
		dcraw.Path = BaseDir + `\utils\dcraw.wasm`
		exiftool.Exec = BaseDir + `\utils\exiftool\exiftool.exe`
		exiftool.Arg1 = strings.TrimSuffix(exiftool.Exec, ".exe")
		exiftool.Config = BaseDir + `\utils\exiftool_config.pl`
	} else {
		ServerMode = name == "rethinkraw-server"
		dcraw.Path = BaseDir + "/utils/dcraw.wasm"
		exiftool.Exec = BaseDir + "/utils/exiftool/exiftool"
		exiftool.Config = BaseDir + "/utils/exiftool_config.pl"
	}

	if testDataDir() == nil {
		return nil
	}
	if data, err := os.UserConfigDir(); err != nil {
		return err
	} else {
		DataDir = filepath.Join(data, "RethinkRAW")
	}
	return testDataDir()
}

func testDataDir() error {
	if err := os.MkdirAll(DataDir, 0700); err != nil {
		return err
	}
	if f, err := os.Create(filepath.Join(DataDir, "lastrun")); err != nil {
		return err
	} else {
		return f.Close()
	}
}
