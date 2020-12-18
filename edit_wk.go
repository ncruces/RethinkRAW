package main

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ncruces/rethinkraw-pkg/osutil"
	"github.com/ncruces/rethinkraw/internal/config"
	"github.com/ncruces/rethinkraw/internal/util"
)

// RethinkRAW edits happen in a workspace.
//
// Adobe DNG Converter loads RAW files and editing metadata from disk,
// and saves DNG files with their embed previews to disk as well.
//
// A workspace is a temporary directory created for each opened RAW file
// where all edits and conversions take place.
//
// This temporary directory is located on: "$TMPDIR/RethingRAW/[HASH]/"
// The directory is created when the file is first opened,
// and deleted when the workspace is finally closed.
//
// It can contain several files:
//  . orig.EXT - a read-only copy of the original RAW file
//  . orig.xmp - a sidecar for orig.EXT
//  . temp.dng - a DNG used as the target for all conversions
//  . edit.dng - a DNG conversion of the original RAW file used for editing previews
//
// Editing settings are loaded from orig.xmp or orig.EXT (in that order).
// The DNG in edit.dng is downscaled to at most 2560 on the widest side.
// When generating a preview, use edit.dng unless the preview requires full resolution.
// If edit.dng is missing, use orig.EXT, ask for a 2560 preview, and save that to edit.dng.

type workspace struct {
	hash      string // a hash of the original RAW file path
	ext       string // the extension of the original RAW file
	base      string // base directory for the workspace
	hasXMP    bool   // did the original RAW file have a XMP sidecar?
	hasPixels bool   // have we extracted pixel data?
	hasEdit   bool   // any recent edits?
}

func openWorkspace(path string) (wk workspace, err error) {
	wk.hash = util.HashedID(filepath.Clean(path))
	wk.ext = filepath.Ext(path)
	wk.base = filepath.Join(config.TempDir, wk.hash) + string(filepath.Separator)

	workspaces.open(wk.hash)
	defer func() {
		if err != nil {
			if workspaces.delete(wk.hash) {
				os.RemoveAll(wk.base)
			}
			wk = workspace{}
		}
	}()

	// Create directory
	err = os.MkdirAll(wk.base, 0700)
	if err != nil {
		return wk, err
	}

	// Have we edited this file recently (10 min)?
	fi, err := os.Stat(wk.base + "edit.dng")
	if err == nil && time.Since(fi.ModTime()) < 10*time.Minute {
		wk.hasEdit = true
		_, e := os.Stat(wk.base + "edit.ppm")
		wk.hasPixels = e == nil
		_, e = os.Stat(wk.base + "orig.xmp")
		wk.hasXMP = e == nil
		return wk, err
	}

	// Was this just copied (1 min)?
	fi, err = os.Stat(wk.base + "orig" + wk.ext)
	if err == nil && time.Since(fi.ModTime()) < time.Minute {
		_, e := os.Stat(wk.base + "orig.xmp")
		wk.hasXMP = e == nil
		return wk, err
	}

	// Otherwise, copy the original RAW file to orig.EXT
	err = osutil.Lnky(path, wk.base+"orig"+wk.ext)
	if err != nil {
		return wk, err
	}

	// And look for a sidecar.
	err = copySidecar(path, wk.base+"orig.xmp")
	if err == nil {
		wk.hasXMP = true
	}
	if os.IsNotExist(err) {
		err = nil
	}
	return wk, err
}

func (wk *workspace) close() {
	if lru := workspaces.close(wk.hash); lru != "" {
		os.RemoveAll(filepath.Join(config.TempDir, lru))
	}
}

// A read-only copy of the original RAW file (full resolution).
func (wk *workspace) orig() string {
	return wk.base + "orig" + wk.ext
}

// A DNG used as the target for all conversions.
func (wk *workspace) temp() string {
	return wk.base + "temp.dng"
}

// A DNG conversion of the original RAW file used for editing previews (downscaled to 2560).
func (wk *workspace) edit() string {
	return wk.base + "edit.dng"
}

// A RAW pixel map for edit.dng.
func (wk *workspace) pixels() string {
	return wk.base + "edit.ppm"
}

// A sidecar for orig.EXT.
func (wk *workspace) origXMP() string {
	return wk.base + "orig.xmp"
}

// The file from which to load editing settings.
func (wk *workspace) loadXMP() string {
	if wk.hasXMP {
		return wk.base + "orig.xmp"
	} else {
		return wk.base + "orig" + wk.ext
	}
}

// HTTP is stateless. There is no notion of a file being opened for editing.
//
// A global manager keeps track of which files are currently being edited,
// and how many tasks are pending for each file.
//
// Once a file has no pending tasks, the workspace is eligible for deletion.
// As an optimization, the 3 LRU workspaces are cached.

var workspaces = workspaceLocker{locks: make(map[string]*workspaceLock)}

const workspaceMaxLRU = 3

type workspaceLocker struct {
	sync.Mutex
	lru   []string
	locks map[string]*workspaceLock
}

type workspaceLock struct {
	sync.Mutex
	n int //
}

// Open and lock a workspace.
func (wl *workspaceLocker) open(hash string) {
	wl.Lock()

	// create a workspace lock
	lk, ok := wl.locks[hash]
	if !ok {
		lk = &workspaceLock{}
		wl.locks[hash] = lk
	}
	lk.n++ // one more pending task

	for i, h := range wl.lru {
		if h == hash {
			// remove workspace from LRU
			wl.lru = append(wl.lru[:i], wl.lru[i+1:]...)
		}
	}

	wl.Unlock()
	lk.Lock()
}

// Close and unlock a workspace, but cache it.
// Return a workspace to evict.
func (wl *workspaceLocker) close(hash string) (lru string) {
	wl.Lock()

	lk := wl.locks[hash]
	lk.n-- // one less pending task

	// are we the last task?
	if lk.n <= 0 {
		// evict a workspace from LRU
		if len(wl.lru) >= workspaceMaxLRU {
			lru, wl.lru = wl.lru[0], wl.lru[1:]
		}
		// add ourselves to LRU
		wl.lru = append(wl.lru, hash)
		// delete our lock
		delete(wl.locks, hash)
	}

	lk.Unlock()
	wl.Unlock()
	return lru // return the evicted workspace
}

// Close and unlock a workspace, but don't cache it.
// Return if safe to delete.
func (wl *workspaceLocker) delete(hash string) (ok bool) {
	wl.Lock()

	lk := wl.locks[hash]
	lk.n-- // one less pending task

	// are we the last task?
	if lk.n <= 0 {
		// delete our lock
		delete(wl.locks, hash)
		ok = true
	}

	lk.Unlock()
	wl.Unlock()
	return ok // were we the last task?
}
