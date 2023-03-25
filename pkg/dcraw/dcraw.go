// Package dcraw provides support running dcraw.
//
// To use this package you need to point it to a WASM/WASI build of dcraw.
// My builds of dcraw are available from:
// https://github.com/ncruces/dcraw
//
// A build of dcraw build can be provided by your application,
// loaded from a file path or embed into your application.
//
// To embed a build of dcraw into your application, import package embed:
//
//	import _ github.com/ncruces/rethinkraw/pkg/dcraw/embed
package dcraw

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"sync"

	_ "embed"

	"github.com/ncruces/go-image/rotateflip"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"golang.org/x/sync/semaphore"
)

// Configure Dcraw.
var (
	Binary []byte // Binary to execute.
	Path   string // Path to load the binary from.
)

var (
	once       sync.Once
	wasm       wazero.Runtime
	module     wazero.CompiledModule
	sem        *semaphore.Weighted
	orienRegex *regexp.Regexp
	thumbRegex *regexp.Regexp
)

func compile() {
	ctx := context.Background()

	wasm = wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, wasm)

	if Binary == nil && Path != "" {
		if bin, err := os.ReadFile(Path); err != nil {
			panic(err)
		} else {
			Binary = bin
		}
	}

	if m, err := wasm.CompileModule(ctx, Binary); err != nil {
		panic(err)
	} else {
		module = m
	}

	sem = semaphore.NewWeighted(6)
	orienRegex = regexp.MustCompile(`Orientation: +(\d)`)
	thumbRegex = regexp.MustCompile(`Thumb size: +(\d+) x (\d+)`)
}

func run(ctx context.Context, root fs.FS, args ...string) ([]byte, error) {
	once.Do(compile)

	err := sem.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer sem.Release(1)

	var buf bytes.Buffer
	cfg := wazero.NewModuleConfig().
		WithArgs(args...).WithStdout(&buf).WithFS(root)
	module, err := wasm.InstantiateModule(ctx, module, cfg)
	if err != nil {
		return nil, err
	}
	err = module.Close(ctx)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetThumb extracts a thumbnail from a RAW file.
//
// The thumbnail will either be a JPEG, or a PNM file in 8-bit P5/P6 format.
// For more about PNM, see https://en.wikipedia.org/wiki/Netpbm
func GetThumb(ctx context.Context, r io.ReadSeeker) ([]byte, error) {
	out, err := run(ctx, readerFS{r}, "dcraw", "-e", "-e", "-c", readerFSname)
	if err != nil {
		return nil, err
	}

	const eoi = "\xff\xd9"
	const tag = "OFfSeTLeNgtH"
	const size = 2 + 2*64/8 + len(tag)
	if off := len(out) - size; off >= 0 &&
		bytes.HasSuffix(out, []byte(tag)) &&
		bytes.HasPrefix(out[off:], []byte(eoi)) {
		offset := int64(binary.LittleEndian.Uint64(out[off+2:]))
		length := int64(binary.LittleEndian.Uint64(out[off+2+64/8:]))
		_, err := r.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, err
		}
		out = append(out[:off], make([]byte, int(length))...)
		_, err = io.ReadFull(r, out[off:])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// GetThumbSize returns the size of the thumbnail [GetThumb] would extract.
// The size is the bigger of width/height, in pixels.
func GetThumbSize(ctx context.Context, r io.ReadSeeker) (int, error) {
	out, err := run(ctx, readerFS{r}, "dcraw", "-i", "-v", readerFSname)
	if err != nil {
		return 0, err
	}

	var max int
	if match := thumbRegex.FindSubmatch(out); match != nil {
		width, _ := strconv.Atoi(string(match[1]))
		height, _ := strconv.Atoi(string(match[2]))
		if width > height {
			max = width
		} else {
			max = height
		}
	}
	return max, nil
}

// GetThumbJPEG extracts a JPEG thumbnail from a RAW file.
//
// This is the same as calling [GetThumb], but converts PNM thumbnails to JPEG.
func GetThumbJPEG(ctx context.Context, r io.ReadSeeker) ([]byte, error) {
	data, err := GetThumb(ctx, r)
	if err != nil {
		return nil, err
	}

	if bytes.HasPrefix(data, []byte("\xff\xd8")) {
		return data, nil
	}

	orientation := make(chan int)
	go func() {
		defer close(orientation)
		orientation <- GetOrientation(ctx, r)
	}()

	img, err := pnmDecodeThumb(data)
	if err != nil {
		return nil, err
	}

	exf := rotateflip.Orientation(<-orientation)
	img = rotateflip.Image(img, exf.Op())

	buf := bytes.Buffer{}
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetOrientation returns the EXIF orientation of the RAW file, or 0 if unknown.
func GetOrientation(ctx context.Context, r io.ReadSeeker) int {
	out, err := run(ctx, readerFS{r}, "dcraw", "-i", "-v", readerFSname)
	if err != nil {
		return 0
	}

	if match := orienRegex.FindSubmatch(out); match != nil {
		return int(match[1][0] - '0')
	}
	return 0
}

// GetRAWPixels develops an half-resolution, demosaiced, not white balanced
// image from the RAW file.
func GetRAWPixels(ctx context.Context, r io.ReadSeeker) ([]byte, error) {
	return run(ctx, readerFS{r}, "dcraw",
		"-r", "1", "1", "1", "1",
		"-o", "0",
		"-h",
		"-4",
		"-t", "0",
		"-c",
		readerFSname)
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
