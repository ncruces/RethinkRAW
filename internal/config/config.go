package config

import (
	"mime"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ncruces/go-exiftool"
	"github.com/ncruces/rethinkraw/pkg/osutil"
)

var (
	BaseDir, DataDir, TempDir string
	Dcraw, DngConverter       string
)

func init() {
	mime.AddExtensionType(".dng", "image/x-adobe-dng")
}

func SetupPaths() (err error) {
	if exe, err := os.Executable(); err != nil {
		return err
	} else {
		BaseDir = filepath.Dir(exe)
	}

	DataDir = filepath.Join(BaseDir, "data")
	TempDir = filepath.Join(os.TempDir(), "RethinkRAW")

	TempDir, err = osutil.GetANSIPath(TempDir)
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		Dcraw = BaseDir + `\utils\dcraw.exe`
		exiftool.Exec = BaseDir + `\utils\exiftool\exiftool.exe`
		exiftool.Arg1 = strings.TrimSuffix(exiftool.Exec, ".exe")
		exiftool.Config = BaseDir + `\utils\exiftool_config.pl`
		DngConverter = os.Getenv("PROGRAMFILES") + `\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
	case "darwin":
		Dcraw = BaseDir + "/utils/dcraw"
		exiftool.Exec = BaseDir + "/utils/exiftool/exiftool"
		exiftool.Config = BaseDir + "/utils/exiftool_config.pl"
		DngConverter = "/Applications/Adobe DNG Converter.app/Contents/MacOS/Adobe DNG Converter"
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
