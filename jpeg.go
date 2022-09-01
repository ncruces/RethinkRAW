package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"

	"github.com/ncruces/go-image/resize"
	"github.com/ncruces/go-image/rotateflip"
	"github.com/ncruces/rethinkraw/internal/dcraw"
)

func previewJPEG(ctx context.Context, path string) ([]byte, error) {
	log.Print("dcraw (get thumb)...")
	data, err := dcraw.GetThumb(ctx, path)
	if err != nil {
		return nil, err
	}

	if bytes.HasPrefix(data, []byte("\xff\xd8")) {
		return data, nil
	}

	img, err := pnmDecodeThumb(data)
	if err != nil {
		return nil, err
	}

	exf := rotateflip.Orientation(rawOrientation(path))
	img = rotateflip.Image(img, exf.Op())

	buf := bytes.Buffer{}
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func exportJPEG(ctx context.Context, path string, settings exportSettings) ([]byte, error) {
	log.Print("dcraw (get thumb)...")
	data, err := dcraw.GetThumb(ctx, path)
	if err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(data, []byte("\xff\xd8")) {
		return nil, errors.New("not a JPEG file")
	}

	if settings.Resample {
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		exf := rotateflip.Orientation(exifOrientation(data))
		img = rotateflip.Image(img, exf.Op())
		fit := settings.FitImage(img.Bounds().Size())
		img = resize.Thumbnail(uint(fit.X), uint(fit.Y), img, resize.Lanczos2)

		buf := bytes.Buffer{}
		// https://fotoforensics.com/tutorial.php?tt=estq
		opt := jpeg.Options{Quality: [13]int{30, 34, 47, 62, 69, 76, 79, 82, 86, 90, 93, 97, 99}[settings.Quality]}
		if err := jpeg.Encode(&buf, img, &opt); err != nil {
			return nil, err
		}

		return append(jfifHeader(settings), buf.Bytes()[2:]...), nil
	}

	return data, err
}

func exifOrientation(data []byte) int {
	if !bytes.HasPrefix(data, []byte("\xff\xd8")) {
		return -1
	}

	data = data[2:]
	for len(data) >= 2 {
		var marker = binary.BigEndian.Uint16(data)

		switch {
		case marker == 0xffff: // padding
			data = data[1:]

		case marker == 0xffe1: // APP1
			if len(data) > 4 {
				size := int(binary.BigEndian.Uint16(data[2:])) + 2
				if 4 <= size && size <= len(data) {
					data = data[4:size]
					if !bytes.HasPrefix(data, []byte("Exif\x00\x00")) {
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

					count := endian.Uint16(data[offset:])
					entries := data[offset+2:]
					if len(data) < 12*int(count) {
						return -2
					} else {
						entries = entries[:12*count]
					}

					for i := 0; i < len(entries); i += 12 {
						tag := endian.Uint16(entries[i:])
						if tag == 0x0112 { // Orientation
							typ := endian.Uint16(entries[i+2:]) // SHORT
							cnt := endian.Uint32(entries[i+4:]) // 1
							val := endian.Uint32(entries[i+8:])
							if typ == 3 && cnt == 1 && val <= 9 {
								return int(val)
							}
							return -2
						}
					}
					return 0
				}
			}
			return -2

		case marker >= 0xffe0: // APPn, JPGn, COM
			if len(data) > 4 {
				size := int(binary.BigEndian.Uint16(data[2:])) + 2
				if 4 <= size && size <= len(data) {
					data = data[size:]
					continue
				}
			}
			return -2

		case marker == 0xff00: // invalid
			return -2

		default:
			return 0
		}
	}
	return -2
}

func jfifHeader(settings exportSettings) []byte {
	if settings.DimUnit == "px" {
		return []byte{'\xff', '\xd8'}
	}

	data := [20]byte{'\xff', '\xd8', '\xff', '\xe0', 0, 16, 'J', 'F', 'I', 'F', 0, 1, 2}
	binary.BigEndian.PutUint16(data[14:], uint16(settings.Density))
	binary.BigEndian.PutUint16(data[16:], uint16(settings.Density))
	if settings.DenUnit == "ppi" {
		data[13] = 1
	} else {
		data[13] = 2
	}
	return data[:]
}

func pnmDecodeThumb(data []byte) (image.Image, error) {
	var format, width, height int
	n, _ := fmt.Fscanf(bytes.NewReader(data), "P%d\n%d %d\n255\n", &format, &width, &height)
	if n == 3 {
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
	return nil, errors.New("unsupported thumbnail")
}
