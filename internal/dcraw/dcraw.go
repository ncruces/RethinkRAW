package dcraw

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/wasi_snapshot_preview1"
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
	thumbRegex *regexp.Regexp
	counter    uint64
)

func compile() {
	ctx := context.Background()

	wasm = wazero.NewRuntime()
	_, err := wasi_snapshot_preview1.Instantiate(ctx, wasm)
	if err != nil {
		panic(err)
	}

	if Binary == nil && Path != "" {
		Binary, err = os.ReadFile(Path)
		if err != nil {
			panic(err)
		}
	}

	module, err = wasm.CompileModule(ctx, Binary, wazero.NewCompileConfig())
	if err != nil {
		panic(err)
	}

	sem = semaphore.NewWeighted(5)

	thumbRegex = regexp.MustCompile(`Thumb size: +(\d+) x (\d+)`)
}

func run(root fs.FS, args ...string) ([]byte, error) {
	once.Do(compile)

	ctx := context.TODO()
	err := sem.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer sem.Release(1)

	var buf bytes.Buffer
	cfg := wazero.NewModuleConfig().
		WithArgs(args...).WithStdout(&buf).WithFS(root).
		WithName("dcraw-" + strconv.FormatUint(atomic.AddUint64(&counter, 1), 10))
	module, err := wasm.InstantiateModule(ctx, module, cfg)
	if err != nil {
		return nil, err
	}
	defer module.Close(ctx)

	return buf.Bytes(), nil
}

func GetThumb(path string) ([]byte, error) {
	return run(fileFS(path), "dcraw", "-e", "-c", "input")
}

func GetThumbSize(path string) (int, error) {
	out, err := run(fileFS(path), "dcraw", "-i", "-v", "input")
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

func GetRAWPixels(path string) ([]byte, error) {
	return run(fileFS(path), "dcraw",
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
	if fs.ValidPath(name) {
		return os.Open(string(file))
	}
	return nil, fs.ErrInvalid
}
