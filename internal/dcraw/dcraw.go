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
	"github.com/tetratelabs/wazero/wasi_snapshot_preview1"
	"golang.org/x/sync/semaphore"
)

//go:embed dcraw.wasm
var Binary []byte

var (
	once       sync.Once
	wasm       wazero.Runtime
	module     wazero.CompiledModule
	sem        *semaphore.Weighted
	thumbRegex *regexp.Regexp
	counter    atomic.Uint64
)

func compile() {
	ctx := context.Background()

	wasm = wazero.NewRuntime(ctx)
	_, err := wasi_snapshot_preview1.Instantiate(ctx, wasm)
	if err != nil {
		panic(err)
	}

	module, err = wasm.CompileModule(ctx, Binary, wazero.NewCompileConfig())
	if err != nil {
		panic(err)
	}

	sem = semaphore.NewWeighted(6)

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

func GetThumb(ctx context.Context, path string) ([]byte, error) {
	out, err := run(ctx, fileFS(path), "dcraw", "-e", "-c", "input")
	if err != nil {
		return nil, err
	}

	if off := len(out) - 20; off >= 0 && bytes.HasPrefix(out[off:], []byte("\xff\xee\x12\x00")) {
		offset := int64(binary.LittleEndian.Uint64(out[off+4+0:]))
		length := int64(binary.LittleEndian.Uint64(out[off+4+8:]))
		f, err := os.Open(path)
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

func GetThumbSize(ctx context.Context, path string) (int, error) {
	out, err := run(ctx, fileFS(path), "dcraw", "-i", "-v", "input")
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

func GetRAWPixels(ctx context.Context, path string) ([]byte, error) {
	return run(ctx, fileFS(path), "dcraw",
		"-r", "1", "1", "1", "1",
		"-o", "0",
		"-h",
		"-4",
		"-t", "0",
		"-c",
		"input")
}

type fileFS string

func (file fileFS) Open(name string) (fs.File, error) {
	if name == "input" {
		return os.Open(string(file))
	}
	if fs.ValidPath(name) {
		return nil, fs.ErrNotExist
	}
	return nil, fs.ErrInvalid
}
