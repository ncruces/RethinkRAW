package main

import (
	"bytes"
	"encoding/xml"
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

	exif "./exiftool"
)

func loadEdit(path string) (xmp xmpSettings, err error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return
	}
	defer wk.close()

	return loadXMP(wk.origXMP())
}

func saveEdit(path string, xmp *xmpSettings) (err error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return
	}
	defer wk.close()

	err = saveXMP(wk.origXMP(), xmp)
	if err != nil {
		return
	}
	wk.hasXMP = true

	dest, err := saveSidecar(path)
	if err != nil {
		return
	}

	if path == dest {
		err = os.RemoveAll(wk.temp())
		if err != nil {
			return
		}

		err = runDNGConverter(wk.orig(), wk.temp(), &exportSettings{DNG: true, Embed: true})
		if err != nil {
			return
		}

		err = lnkyFile(wk.temp(), dest+".bak")
	} else {
		err = copyFile(wk.origXMP(), dest+".bak")
	}

	if err != nil {
		return
	}
	return os.Rename(dest+".bak", dest)
}

func previewEdit(path string, xmp *xmpSettings) (thumb []byte, err error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return
	}
	defer wk.close()

	err = saveXMP(wk.lastXMP(), xmp)
	if err != nil {
		return
	}

	err = os.RemoveAll(wk.temp())
	if err != nil {
		return
	}

	err = runDNGConverter(wk.last(), wk.temp(), nil)
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

	err = saveXMP(wk.origXMP(), xmp)
	if err != nil {
		return
	}
	wk.hasXMP = true

	err = os.RemoveAll(wk.temp())
	if err != nil {
		return
	}

	var writer io.WriteCloser
	var result *exif.AsyncResult
	if !(exp.DNG || exp.Resample) {
		writer, result = fixMetaJPEGAsync(wk.orig())
	}

	err = runDNGConverter(wk.orig(), wk.temp(), exp)
	if err != nil {
		return
	}

	if exp.DNG {
		err = fixMetaDNG(wk.orig(), wk.temp(), path)
		if err != nil {
			return
		}

		return ioutil.ReadFile(wk.temp())
	} else {
		data, err = exportJPEG(wk.temp(), exp)
		if err != nil || exp.Resample {
			return
		}

		writer.Write(data)
		writer.Close()
		return result.Get()
	}
}

func exportHeaders(path string, exp *exportSettings, headers http.Header) {
	var ext string
	if exp.DNG {
		ext = ".dng"
	} else {
		ext = ".jpg"
	}
	attachmentHeaders(path, ext, headers)
}

type exportSettings struct {
	DNG      bool
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
	hasXMP  bool
	hasEdit bool
}

func openWorkspace(path string) (wk workspace, err error) {
	wk.hash = md5sum(filepath.Clean(path))
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
		return
	}

	fi, err := os.Stat(wk.base + "edit.dng")
	if err == nil && time.Since(fi.ModTime()) < 10*time.Minute {
		_, e := os.Stat(wk.base + "orig.xmp")
		wk.hasXMP = e == nil
		wk.hasEdit = true
		return
	}

	fi, err = os.Stat(wk.base + "orig" + wk.ext)
	if err == nil && time.Since(fi.ModTime()) < time.Minute {
		_, e := os.Stat(wk.base + "orig.xmp")
		wk.hasXMP = e == nil
		return
	}

	err = lnkyFile(path, wk.base+"orig"+wk.ext)
	if err != nil {
		return
	}

	err = copySidecar(path, wk.base+"orig.xmp")
	if os.IsNotExist(err) {
		err = nil
	} else if err == nil {
		wk.hasXMP = true
	}
	return
}

func (wk *workspace) close() {
	if old := workspaces.close(wk.hash); old != "" {
		os.RemoveAll(filepath.Join(tempDir, old))
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
	if wk.hasXMP {
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

func (wk *workspace) lastXMP() string {
	if wk.hasEdit {
		return wk.edit()
	} else {
		return wk.origXMP()
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

func (wl *workspaceLocker) close(hash string) (old string) {
	wl.Lock()

	lk := wl.locks[hash]
	lk.n--

	if lk.n <= 0 {
		if len(wl.lru) >= workspaceMaxLRU {
			old, wl.lru = wl.lru[0], wl.lru[1:]
		}
		wl.lru = append(wl.lru, hash)
		delete(wl.locks, hash)
	}

	lk.Unlock()
	wl.Unlock()
	return
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
	return
}

func copySidecar(src, dst string) error {
	var d []byte
	err := os.ErrNotExist
	ext := filepath.Ext(src)

	if ext != "" {
		// if NAME.XMP is there for NAME.EXT, use it
		xmp := strings.TrimSuffix(src, ext) + ".xmp"
		d, err = ioutil.ReadFile(xmp)
		if err == nil && !isSidecarForExt(d, ext) {
			err = os.ErrNotExist
		}
	}
	if os.IsNotExist(err) {
		// if NAME.EXT.XMP is there for NAME.EXT, use it
		d, err = ioutil.ReadFile(src + ".xmp")
		if err == nil && !isSidecarForExt(d, ext) {
			err = os.ErrNotExist
		}
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, d, 0600)
}

func saveSidecar(src string) (string, error) {
	// fallback to NAME.EXT.XMP
	ret := src + ".xmp"
	ext := filepath.Ext(src)

	if ext != "" {
		// if NAME.XMP is there for NAME.EXT, use it
		xmp := strings.TrimSuffix(src, ext) + ".xmp"
		if d, err := ioutil.ReadFile(xmp); err == nil {
			if isSidecarForExt(d, ext) {
				return xmp, nil
			}
		} else if !os.IsNotExist(err) {
			return "", err
		} else {
			// fallback to NAME.XMP
			ret = xmp
		}
	}

	// if NAME.EXT.XMP exists, use it
	if _, err := os.Stat(src + ".xmp"); err == nil {
		return src + ".xmp", nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	// if NAME.DNG has embed XMP, use it
	if strings.EqualFold(ext, ".dng") && hasXMP(src) {
		return src, nil
	}

	// otherwise, fallback
	return ret, nil
}

func isSidecarForExt(data []byte, ext string) bool {
	dec := xml.NewDecoder(bytes.NewBuffer(data))
	for {
		t, err := dec.Token()
		if err != nil {
			return err == io.EOF
		}

		if s, ok := t.(xml.StartElement); ok {
			for _, a := range s.Attr {
				if a.Name.Local == "SidecarForExtension" &&
					strings.HasPrefix(a.Name.Space, "http://ns.adobe.com/photoshop/") {
					return strings.EqualFold(ext, "."+a.Value)
				}
			}
		}
	}
}
