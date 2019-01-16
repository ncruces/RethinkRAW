package main

import (
	"image"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func loadEdit(path string) (xmp xmpSettings, err error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return
	}
	defer wk.close()

	return loadXmp(wk.origXmp())
}

func previewEdit(path string, xmp *xmpSettings) (thumb []byte, err error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return
	}
	defer wk.close()

	err = saveXmp(wk.lastXmp(), xmp)
	if err != nil {
		return
	}

	err = os.RemoveAll(wk.temp())
	if err != nil {
		return
	}

	err = toDng(wk.last(), wk.temp(), nil)
	if err != nil {
		return
	}

	err = os.Rename(wk.temp(), wk.edit())
	if err != nil {
		return
	}

	return previewJPEG(wk.edit())
}

func exportEdit(path string, xmp *xmpSettings, exp *exportSettings) (data []byte, err error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return
	}
	defer wk.close()

	err = saveXmp(wk.origXmp(), xmp)
	if err != nil {
		return
	}

	err = os.RemoveAll(wk.temp())
	if err != nil {
		return
	}

	err = toDng(wk.orig(), wk.temp(), exp)
	if err != nil {
		return
	}

	if exp.Dng {
		err = copyMeta(wk.orig(), wk.temp(), path)
		if err != nil {
			return
		}

		return ioutil.ReadFile(wk.temp())
	} else {
		return exportJPEG(wk.temp(), exp)
	}
}

func exportHeaders(path string, exp *exportSettings, headers http.Header) {
	var name, mime string

	name = filename(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	if name == "" {
		name = "photo"
	}
	if exp.Dng {
		mime = "image/x-adobe-dng"
		name += ".dng"
	} else {
		mime = "image/jpeg"
		name += ".jpg"
	}

	headers.Add("Content-Disposition", `attachment; filename="`+name+`"`)
	headers.Add("Content-Type", mime)
}

type exportSettings struct {
	Dng      bool
	Preview  string
	FastLoad bool
	Embed    bool
	Lossy    bool

	Resample bool
	Quality  int
	Fit      string
	Long     float64
	Short    float64
	Width    float64
	Height   float64
	DimUnit  string
	Density  int
	DenUnit  string
	MPixels  float64
}

func (ex *exportSettings) FitImage(size image.Point) (fit image.Point) {
	if ex.Fit == "mpix" {
		mul := math.Sqrt(1e6 * ex.MPixels / float64(size.X*size.Y))
		if size.X > size.Y {
			fit.X = MaxInt
			fit.Y = int(mul * float64(size.Y))
		} else {
			fit.X = int(mul * float64(size.X))
			fit.Y = MaxInt
		}
	} else {
		mul := 1.0

		if ex.DimUnit != "px" {
			density := float64(ex.Density)
			if ex.DimUnit == "in" {
				if ex.DenUnit == "ppi" {
					mul = density
				} else {
					mul = density * 2.54
				}
			} else {
				if ex.DenUnit == "ppi" {
					mul = density / 2.54
				} else {
					mul = density
				}
			}
		}

		round := func(x float64) int {
			i := int(x + 0.5)
			if i > 0 {
				return i
			}
			return MaxInt
		}

		if ex.Fit == "dims" {
			long, short := ex.Long, ex.Short
			if 0 < long && long < short {
				long, short = short, long
			}

			if size.X > size.Y {
				fit.X = round(mul * long)
				fit.Y = round(mul * short)
			} else {
				fit.X = round(mul * long)
				fit.Y = round(mul * short)
			}
		} else {
			fit.X = round(mul * ex.Width)
			fit.Y = round(mul * ex.Height)
		}
	}
	return
}

type workspace struct {
	hash    string
	base    string
	ext     string
	hasXmp  bool
	hasEdit bool
	mutex   sync.Mutex
}

var workspaces = make(map[string]*workspace)
var workspacesMutex sync.Mutex

const maxWorkspaces = 5

func openWorkspace(path string) (wk *workspace, err error) {
	workspacesMutex.Lock()
	defer workspacesMutex.Unlock()

	path = filepath.Clean(path)
	hash := md5sum(filepath.ToSlash(path))

	wk, ok := workspaces[hash]
	if !ok {
		wk = &workspace{hash: hash}
		wk.ext = filepath.Ext(path)
		wk.base = filepath.Join(tempDir, hash) + string(filepath.Separator)
	}

	err = os.MkdirAll(wk.base, 0700)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			os.RemoveAll(wk.base)
			wk = nil
		} else {
			wk.mutex.Lock()
			delete(workspaces, hash)
		}
	}()

	fi, err := os.Stat(wk.base + "edit.dng")
	if err == nil && time.Since(fi.ModTime()) < 10*time.Minute {
		_, e := os.Stat(wk.base + "orig.xmp")
		wk.hasXmp = e == nil
		wk.hasEdit = true
		return
	}

	err = copyFile(path, wk.base+"orig"+wk.ext)
	if err != nil {
		return
	}

	path = strings.TrimSuffix(path, wk.ext) + ".xmp"

	err = copyFile(path, wk.base+"orig.xmp")
	if os.IsNotExist(err) {
		err = nil
	} else if err == nil {
		wk.hasXmp = true
	}
	return
}

func (wk *workspace) close() {
	workspacesMutex.Lock()
	defer workspacesMutex.Unlock()

	if len(workspaces) >= maxWorkspaces {
		for k, w := range workspaces {
			delete(workspaces, k)
			os.RemoveAll(w.base)
			break
		}
	}

	workspaces[wk.hash] = wk
	wk.mutex.Unlock()
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

func (wk *workspace) origXmp() string {
	if wk.hasXmp {
		return wk.base + "orig.xmp"
	} else {
		return wk.base + "orig" + wk.ext
	}
}

func (wk *workspace) last() string {
	if wk.hasEdit {
		return wk.edit()
	} else {
		return wk.orig()
	}
}

func (wk *workspace) lastXmp() string {
	if wk.hasEdit {
		return wk.edit()
	} else {
		return wk.origXmp()
	}
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, in)
	return
}
