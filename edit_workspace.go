package main

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"rethinkraw/osutil"
)

type workspace struct {
	hash    string
	base    string
	ext     string
	hasXMP  bool
	hasEdit bool
}

func openWorkspace(path string) (wk workspace, err error) {
	wk.hash = hash(filepath.Clean(path))
	wk.base = filepath.Join(tempDir, wk.hash) + string(filepath.Separator)
	wk.ext = filepath.Ext(path)

	workspaces.open(wk.hash)
	defer func() {
		if err != nil {
			if workspaces.delete(wk.hash) {
				os.RemoveAll(wk.base)
			}
			wk = workspace{}
		}
	}()

	err = os.MkdirAll(wk.base, 0700)
	if err != nil {
		return wk, err
	}

	fi, err := os.Stat(wk.base + "edit.dng")
	if err == nil && time.Since(fi.ModTime()) < 10*time.Minute {
		_, e := os.Stat(wk.base + "orig.xmp")
		wk.hasXMP = e == nil
		wk.hasEdit = true
		return wk, err
	}

	fi, err = os.Stat(wk.base + "orig" + wk.ext)
	if err == nil && time.Since(fi.ModTime()) < time.Minute {
		_, e := os.Stat(wk.base + "orig.xmp")
		wk.hasXMP = e == nil
		return wk, err
	}

	sfi, err := os.Stat(path)
	if err != nil {
		return wk, err
	}

	if os.SameFile(fi, sfi) {
		_, e := os.Stat(wk.base + "orig.xmp")
		wk.hasXMP = e == nil
		return wk, err
	}

	err = osutil.Lnky(path, wk.base+"orig"+wk.ext)
	if err != nil {
		return wk, err
	}

	err = loadSidecar(path, wk.base+"orig.xmp")
	if os.IsNotExist(err) {
		err = nil
	} else if err == nil {
		wk.hasXMP = true
	}
	return wk, err
}

func (wk *workspace) close() {
	if lru := workspaces.close(wk.hash); lru != "" {
		os.RemoveAll(filepath.Join(tempDir, lru))
	}
}

func (wk *workspace) orig() string {
	return wk.base + "orig" + wk.ext
}

func (wk *workspace) edit() string {
	return wk.base + "edit.dng"
}

func (wk *workspace) temp() string {
	return wk.base + "temp.dng"
}

func (wk *workspace) origXMP() string {
	return wk.base + "orig.xmp"
}

func (wk *workspace) loadXMP() string {
	if wk.hasXMP {
		return wk.base + "orig.xmp"
	} else {
		return wk.base + "orig" + wk.ext
	}
}

func (wk *workspace) last() string {
	if wk.hasEdit {
		return wk.base + "edit.dng"
	} else {
		return wk.base + "orig" + wk.ext
	}
}

func (wk *workspace) lastXMP() string {
	if wk.hasEdit {
		return wk.base + "edit.dng"
	} else {
		return wk.base + "orig.xmp"
	}
}

type workspaceLock struct {
	sync.Mutex
	n int
}

type workspaceLocker struct {
	sync.Mutex
	lru   []string
	locks map[string]*workspaceLock
}

var workspaces = workspaceLocker{locks: make(map[string]*workspaceLock)}

const workspaceMaxLRU = 3

func (wl *workspaceLocker) open(hash string) {
	wl.Lock()

	lk, ok := wl.locks[hash]
	if !ok {
		lk = &workspaceLock{}
		wl.locks[hash] = lk
	}
	lk.n++

	for i, h := range wl.lru {
		if h == hash {
			wl.lru = append(wl.lru[:i], wl.lru[i+1:]...)
		}
	}

	wl.Unlock()
	lk.Lock()
}

func (wl *workspaceLocker) close(hash string) (lru string) {
	wl.Lock()

	lk := wl.locks[hash]
	lk.n--

	if lk.n <= 0 {
		if len(wl.lru) >= workspaceMaxLRU {
			lru, wl.lru = wl.lru[0], wl.lru[1:]
		}
		wl.lru = append(wl.lru, hash)
		delete(wl.locks, hash)
	}

	lk.Unlock()
	wl.Unlock()
	return lru
}

func (wl *workspaceLocker) delete(hash string) (ok bool) {
	wl.Lock()

	lk := wl.locks[hash]
	lk.n--

	if lk.n <= 0 {
		delete(wl.locks, hash)
		ok = true
	}

	lk.Unlock()
	wl.Unlock()
	return ok
}
