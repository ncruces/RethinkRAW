package dng

import (
	"bytes"
	"errors"
	"io/ioutil"

	"github.com/rwcarlsen/goexif/tiff"
)

func GetDCPProfileName(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	if len(data) > 3 {
		if data[0] == 'I' {
			data[2] = 42
			data[3] = 0
		} else {
			data[2] = 0
			data[3] = 42
		}
	}

	tif, err := tiff.Decode(bytes.NewBuffer(data))
	if len(tif.Dirs) > 0 {
		for _, tag := range tif.Dirs[0].Tags {
			if tag.Id == 0xc6f8 { // ProfileName
				return tag.StringVal()
			}
		}
	}
	return "", errors.New("dcp: could not find ProfileName tag")
}
