package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os/exec"
	"syscall"
)

var dcraw = `.\utils\dcraw.exe`
var jpegtran = `.\utils\jpegtran.exe`

type constError string

func (e constError) Error() string { return string(e) }

const unsupportedThumb = constError("unsupported thumbnail")

func getJpeg(path string) ([]byte, error) {
	log.Printf("dcraw [-e -c %s]\n", path)
	cmd := exec.Command(dcraw, "-e", "-c", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	if len(out) > 2 && out[0] == '\xff' && out[1] == '\xd8' {
		return out, nil
	}

	img, err := pnmDecodeThumb(out)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func getThumb(path string) ([]byte, error) {
	data, err := getJpeg(path)
	if err != nil {
		return nil, err
	}

	exif := exifOrientation(data)
	switch {
	case exif == -1:
		return nil, errors.New("not a JPEG file")
	case exif < 0 || exif > 9:
		return nil, errors.New("not a valid JPEG file")
	case exif < 2 || exif == 9:
		return data, nil
	}

	flags := [7][]string{
		{"-flip", "horizontal"},
		{"-rotate", "180"},
		{"-flip", "vertical"},
		{"-transpose"},
		{"-rotate", "90"},
		{"-transverse"},
		{"-rotate", "270"},
	}

	opts := append([]string{"-trim", "-copy", "none"}, flags[exif-2]...)

	log.Printf("jpegtran %v\n", opts)
	cmd := exec.Command(jpegtran, opts...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdin = bytes.NewReader(data)
	return cmd.Output()
}

func exifOrientation(data []byte) int {
	if len(data) < 2 || data[0] != '\xff' || data[1] != '\xd8' {
		return -1
	}

	data = data[2:]
	for len(data) >= 2 {
		var marker = binary.BigEndian.Uint16(data)

		switch {
		case marker == 0xffff:
			data = data[1:]

		case marker == 0xffe1:
			if len(data) > 4 {
				size := int(binary.BigEndian.Uint16(data[2:])) + 2
				if 4 <= size && size <= len(data) {
					data = data[4:size]
					if len(data) < 6 || string(data[0:6]) != "Exif\x00\x00" {
						return 0
					}

					data = data[6:]
					if len(data) < 8 {
						return -2
					}

					var endian binary.ByteOrder
					switch string(data[0:4]) {
					case "II*\x00":
						endian = binary.LittleEndian
					case "MM\x00*":
						endian = binary.BigEndian
					default:
						return -2
					}

					offset := endian.Uint32(data[4:])
					if len(data) < int(offset)+2 {
						return -2
					}

					tags := int(endian.Uint16(data[offset:]))
					data = data[offset+2:]

					if len(data) < 12*tags {
						return -2
					}

					for i := 0; i < tags; i++ {
						if endian.Uint16(data[i*12:]) == 0x0112 {
							v := int(endian.Uint16(data[i*12+8:]))
							if v > 9 {
								v = -2
							}
							return v
						}
					}
					return 0
				}
			}
			return -2

		case marker >= 0xffe0:
			if len(data) > 4 {
				size := int(binary.BigEndian.Uint16(data[2:])) + 2
				if 4 <= size && size <= len(data) {
					data = data[size:]
					continue
				}
			}
			return -2

		case marker == 0xff00:
			return -2

		default:
			return 0
		}
	}
	return -2
}

func pnmDecodeThumb(data []byte) (image.Image, error) {
	var format, width, height int
	n, err := fmt.Sscanf(string(data), "P%d\n%d %d\n255\n", &format, &width, &height)
	if err == nil && n == 3 {
		for i := 0; i < 3; i++ {
			data = data[bytes.IndexByte(data, '\n')+1:]
		}

		rect := image.Rect(0, 0, width, height)
		switch {
		case format == 5 && len(data) == width*height:
			img := image.NewGray(rect)
			copy(img.Pix, data)
			return img, nil

		case format == 6 && len(data) == 3*width*height:
			if len(data) != 3*width*height {
				return nil, unsupportedThumb
			}
			img := image.NewRGBA(rect)
			var i, j int
			for k := 0; k < width*height; k++ {
				img.Pix[i+0] = data[j+0]
				img.Pix[i+1] = data[j+1]
				img.Pix[i+2] = data[j+2]
				img.Pix[i+3] = 255
				i += 4
				j += 3
			}
			return img, nil
		}
	}

	return nil, unsupportedThumb
}
