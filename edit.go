package main

import (
	"bytes"
	"encoding/xml"
	"image"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"rethinkraw/osutil"
)

func loadEdit(path string) (xmp xmpSettings, err error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return xmp, err
	}
	defer wk.close()

	return loadXMP(wk.loadXMP())
}

func saveEdit(path string, xmp *xmpSettings) error {
	wk, err := openWorkspace(path)
	if err != nil {
		return err
	}
	defer wk.close()

	err = editXMP(wk.origXMP(), xmp)
	if err != nil {
		return err
	}

	dest, err := destSidecar(path)
	if err != nil {
		return err
	}

	if path == dest {
		err = os.RemoveAll(wk.temp())
		if err != nil {
			return err
		}

		exp := exportSettings{
			DNG:     true,
			Embed:   true,
			Preview: dngPreview(wk.orig()),
		}

		err = runDNGConverter(wk.orig(), wk.temp(), 0, &exp)
		if err != nil {
			return err
		}

		err = osutil.Lnky(wk.temp(), dest+".bak")
	} else {
		err = osutil.Copy(wk.origXMP(), dest+".bak")
	}

	if err != nil {
		return err
	}
	return os.Rename(dest+".bak", dest)
}

func previewEdit(path string, xmp *xmpSettings, pvw *previewSettings) ([]byte, error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return nil, err
	}
	defer wk.close()

	if pvw.Preview > 0 {
		err = editXMP(wk.lastXMP(), xmp)
		if err != nil {
			return nil, err
		}

		err = os.RemoveAll(wk.temp())
		if err != nil {
			return nil, err
		}

		var side int
		if wk.hasEdit {
			side = pvw.Preview
		} else {
			side = 2560
		}

		err = runDNGConverter(wk.last(), wk.temp(), side, nil)
		if err != nil {
			return nil, err
		}

		if wk.hasEdit {
			return previewJPEG(wk.temp())
		} else {
			err = os.Rename(wk.temp(), wk.edit())
			if err != nil {
				return nil, err
			}

			return previewJPEG(wk.edit())
		}
	} else {
		err = editXMP(wk.origXMP(), xmp)
		if err != nil {
			return nil, err
		}

		err = os.RemoveAll(wk.temp())
		if err != nil {
			return nil, err
		}

		err = runDNGConverter(wk.orig(), wk.temp(), 0, &exportSettings{})
		if err != nil {
			return nil, err
		}

		return previewJPEG(wk.temp())
	}
}

func exportEdit(path string, xmp *xmpSettings, exp *exportSettings) ([]byte, error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return nil, err
	}
	defer wk.close()

	err = editXMP(wk.origXMP(), xmp)
	if err != nil {
		return nil, err
	}

	err = os.RemoveAll(wk.temp())
	if err != nil {
		return nil, err
	}

	var reader io.ReadCloser
	var writer io.WriteCloser
	if !exp.DNG && !exp.Resample {
		writer, reader, err = fixMetaJPEGAsync(wk.orig())
		if err != nil {
			return nil, err
		}
	}

	err = runDNGConverter(wk.orig(), wk.temp(), 0, exp)
	if err != nil {
		return nil, err
	}

	if exp.DNG {
		err = fixMetaDNG(wk.orig(), wk.temp(), path)
		if err != nil {
			return nil, err
		}

		return ioutil.ReadFile(wk.temp())
	} else {
		data, err := exportJPEG(wk.temp(), exp)
		if err != nil || exp.Resample {
			return data, err
		}

		go func() {
			writer.Write(data)
			writer.Close()
		}()
		return ioutil.ReadAll(reader)
	}
}

func exportName(path string, exp *exportSettings) string {
	var ext string
	if exp.DNG {
		ext = ".dng"
	} else {
		ext = ".jpg"
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) + ext
}

type previewSettings struct {
	Preview int
}

type exportSettings struct {
	DNG     bool
	Preview string
	Lossy   bool
	Embed   bool

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
	return fit
}

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

func loadSidecar(src, dst string) error {
	var d []byte
	err := os.ErrNotExist
	ext := filepath.Ext(src)

	if ext != "" {
		// if NAME.XMP is there for NAME.EXT, use it
		xmp := strings.TrimSuffix(src, ext) + ".xmp"
		d, err = ioutil.ReadFile(xmp)
		if err == nil && !isSidecarForExt(bytes.NewReader(d), ext) {
			err = os.ErrNotExist
		}
	}
	if os.IsNotExist(err) {
		// if NAME.EXT.XMP is there for NAME.EXT, use it
		d, err = ioutil.ReadFile(src + ".xmp")
		if err == nil && !isSidecarForExt(bytes.NewReader(d), ext) {
			err = os.ErrNotExist
		}
	}
	if os.IsNotExist(err) {
		// extract embed XMP
		return extractXMP(src, dst)
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, d, 0666)
}

func destSidecar(src string) (string, error) {
	// fallback to NAME.EXT.XMP
	ret := src + ".xmp"
	ext := filepath.Ext(src)

	if ext != "" {
		// if NAME.XMP is there for NAME.EXT, use it
		xmp := strings.TrimSuffix(src, ext) + ".xmp"
		if f, err := os.Open(xmp); err == nil {
			defer f.Close()
			if isSidecarForExt(f, ext) {
				return xmp, nil
			}
			f.Close()
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

	// if NAME.DNG was edited, use it
	if strings.EqualFold(ext, ".dng") && hasEdits(src) {
		return src, nil
	}

	// otherwise, fallback
	return ret, nil
}

func isSidecarForExt(r io.Reader, ext string) bool {
	test := func(name xml.Name) bool {
		return name.Local == "SidecarForExtension" &&
			(name.Space == "http://ns.adobe.com/photoshop/1.0/" || name.Space == "photoshop")
	}

	dec := xml.NewDecoder(r)
	for {
		t, err := dec.Token()
		if err != nil {
			return err == io.EOF
		}

		if s, ok := t.(xml.StartElement); ok {
			if test(s.Name) {
				t, _ := dec.Token()
				v, ok := t.(xml.CharData)
				return ok && strings.EqualFold(ext, "."+string(v))
			}
			for _, a := range s.Attr {
				if test(a.Name) {
					return strings.EqualFold(ext, "."+a.Value)
				}
			}
		}
	}
}
