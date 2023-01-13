package dcraw

import (
	"io"
	"io/fs"
	"time"
)

// These implement an [fs.FS] with a single root directory,
// and a single file in that directory, named [readerFSname],
// that reads from the [io.ReadSeeker].

type readerFS struct{ io.ReadSeeker }
type readerDir struct{ io.ReadSeeker }

const readerFSname = "input"

func (f readerFS) Open(name string) (fs.File, error) {
	if name == "." {
		return readerDir{}, nil
	}
	if name == readerFSname {
		_, err := f.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
	if fs.ValidPath(name) {
		return nil, fs.ErrNotExist
	}
	return nil, fs.ErrInvalid
}

func (f readerFS) Close() error { return nil }

func (f readerFS) Stat() (fs.FileInfo, error) { return readerFS{f}, nil }

func (f readerFS) Name() string {
	if f.ReadSeeker == nil {
		return "."
	}
	return readerFSname
}

func (f readerFS) Size() int64 {
	if f.ReadSeeker == nil {
		return 0
	}
	current, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0
	}
	end, _ := f.Seek(0, io.SeekEnd)
	f.Seek(current, io.SeekStart)
	return end
}

func (f readerFS) Mode() fs.FileMode {
	if f.ReadSeeker == nil {
		return fs.ModeDir | 0500
	}
	return 0400
}

func (f readerFS) Type() fs.FileMode { return f.Mode().Type() }

func (f readerFS) ModTime() time.Time { return time.Time{} }

func (f readerFS) IsDir() bool { return f.ReadSeeker == nil }

func (f readerFS) Sys() any { return nil }

func (f readerFS) Info() (fs.FileInfo, error) { return f, nil }

func (d readerDir) Close() error { return nil }

func (d readerDir) Read([]byte) (int, error) { return 0, nil }

func (d readerDir) Stat() (fs.FileInfo, error) { return readerFS{nil}, nil }

func (d readerDir) ReadDir(n int) ([]fs.DirEntry, error) {
	return []fs.DirEntry{readerFS{d}, nil}, nil
}
