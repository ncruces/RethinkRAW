package dngconv

import "os"

func CheckInstall() error {
	const converter = "/Applications/Adobe DNG Converter.app/Contents/MacOS/Adobe DNG Converter"
	_, err := os.Stat(converter)
	if err != nil {
		return err
	}
	conv = converter
	return nil
}

func dngPath(path string) (string, error) {
	return path, nil
}
