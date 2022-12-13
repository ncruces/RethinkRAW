package dngconv

import (
	"os"

	"github.com/ncruces/rethinkraw/pkg/osutil"
)

func CheckInstall() error {
	converter := os.Getenv("PROGRAMFILES") + `\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe`
	_, err := os.Stat(converter)
	if err != nil {
		return err
	}
	conv = converter
	return nil
}

func dngPath(path string) (string, error) {
	return osutil.GetANSIPath(path)
}
