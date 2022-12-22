// Package dcraw provides support running an embed version of dcraw.
//
// Importing this package embeds a WASM build of dcraw into your binaries.
// Source code for that build of dcraw is available from:
// https://github.com/ncruces/dcraw/blob/ncruces-rethinkraw/dcraw.c
package dcraw

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"

	_ "embed"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"golang.org/x/sync/semaphore"
)

//go:embed dcraw.wasm
var Binary []byte

var (
	once       sync.Once
	wasm       wazero.Runtime
	module     wazero.CompiledModule
	sem        *semaphore.Weighted
	orienRegex *regexp.Regexp
	thumbRegex *regexp.Regexp
	counter    atomic.Uint64
)

func compile() {
	ctx := context.Background()

	wasm = wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, wasm)
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
		WithArgs(args...).WithStdout(&buf).WithFS(root).
		WithName("dcraw-" + strconv.FormatUint(counter.Add(1), 10))
	module, err := wasm.InstantiateModule(ctx, module, cfg)
	if err != nil {
		return nil, err
	}
	defer module.Close(ctx)

	return buf.Bytes(), nil
}

// GetThumb extracts a thumbnail from a RAW file.
//
// The thumbnail will either be a JPEG, or a PNM file in 8-bit P5/P6 format.
// For more about PNM, see https://en.wikipedia.org/wiki/Netpbm
func GetThumb(ctx context.Context, file string) ([]byte, error) {
	out, err := run(ctx, fileFS(file), "dcraw", "-e", "-e", "-c", fileFSname)
	if err != nil {
		return nil, err
	}

	if off := len(out) - 20; off >= 0 && bytes.HasPrefix(out[off:], []byte("\xff\xee\x12\x00")) {
		offset := int64(binary.LittleEndian.Uint64(out[off+4+0:]))
		length := int64(binary.LittleEndian.Uint64(out[off+4+8:]))
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		out = append(out[:off], make([]byte, int(length))...)
		_, err = io.ReadFull(io.NewSectionReader(f, offset, length), out[off:])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// GetThumbSize returns the size of the thumbnail [GetThumb] would extract.
// The size is the bigger of width/height, in pixels.
func GetThumbSize(ctx context.Context, file string) (int, error) {
	out, err := run(ctx, fileFS(file), "dcraw", "-i", "-v", fileFSname)
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

// GetOrientation returns the EXIF orientation of the RAW file, or 0 if unknown.
func GetOrientation(ctx context.Context, file string) int {
	out, err := run(ctx, fileFS(file), "dcraw", "-i", "-v", fileFSname)
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
func GetRAWPixels(ctx context.Context, file string) ([]byte, error) {
	return run(ctx, fileFS(file), "dcraw",
		"-r", "1", "1", "1", "1",
		"-o", "0",
		"-h",
		"-4",
		"-t", "0",
		"-c",
		fileFSname)
}

type fileFS string

const fileFSname = "input"

func (file fileFS) Open(name string) (fs.File, error) {
	if name == fileFSname {
		return os.Open(string(file))
	}
	if fs.ValidPath(name) {
		return nil, fs.ErrNotExist
	}
	return nil, fs.ErrInvalid
}
