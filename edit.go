package main

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func loadEdit(path string) (xmp xmpSettings, err error) {
	wk, err := newWorkspace(path)
	if err != nil {
		return
	}

	return loadXmp(wk.OrigXmp())
}

func previewEdit(path string, xmp *xmpSettings) (thumb []byte, err error) {
	wk, err := newWorkspace(path)
	if err != nil {
		return
	}

	err = saveXmp(wk.LastXmp(), xmp)
	if err != nil {
		return
	}

	err = os.RemoveAll(wk.Temp())
	if err != nil {
		return
	}

	err = toDng(wk.Last(), filepath.Base(wk.Temp()), filepath.Dir(wk.Temp()), true, true)
	if err != nil {
		return
	}

	err = os.Rename(wk.Temp(), wk.Edit())
	if err != nil {
		return
	}

	return getThumb(wk.Edit())
}

func exportEdit(path string, xmp *xmpSettings, exp *exportSettings) (data []byte, err error) {
	wk, err := newWorkspace(path)
	if err != nil {
		return
	}

	err = saveXmp(wk.OrigXmp(), xmp)
	if err != nil {
		return
	}

	err = os.RemoveAll(wk.Temp())
	if err != nil {
		return
	}

	err = toDng(wk.Orig(), filepath.Base(wk.Temp()), filepath.Dir(wk.Temp()), false, false)
	if err != nil {
		return
	}

	err = os.Rename(wk.Temp(), wk.Edit())
	if err != nil {
		return
	}

	return getJpeg(wk.Edit())
}

type exportSettings struct {
	Resample bool
	Quality  int
	Fit      string
	Long     float32
	Short    float32
	Width    float32
	Height   float32
	DimUnit  string
	Density  int
	DenUnit  string
	MPixels  float32
}

type workspace struct {
	base    string
	ext     string
	hasXmp  bool
	hasEdit bool
}

func (wk *workspace) Orig() string {
	return wk.base + "orig" + wk.ext
}

func (wk *workspace) Edit() string {
	return wk.base + "edit.dng"
}

func (wk *workspace) Temp() string {
	return wk.base + "temp.dng"
}

func (wk *workspace) OrigXmp() string {
	if wk.hasXmp {
		return wk.base + "orig.xmp"
	} else {
		return wk.base + "orig" + wk.ext
	}
}

func (wk *workspace) Last() string {
	if wk.hasEdit {
		return wk.Edit()
	} else {
		return wk.Orig()
	}
}

func (wk *workspace) LastXmp() string {
	if wk.hasEdit {
		return wk.Edit()
	} else {
		return wk.OrigXmp()
	}
}

func newWorkspace(path string) (wk workspace, err error) {
	path = filepath.Clean(path)
	hash := hash(filepath.ToSlash(path))

	wk.base = filepath.Join(tempDir, hash) + string(filepath.Separator)
	wk.ext = filepath.Ext(path)

	err = os.MkdirAll(wk.base, 700)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			os.RemoveAll(wk.base)
			wk = workspace{}
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

func hash(data string) string {
	h := md5.Sum([]byte(data))
	return base64.URLEncoding.EncodeToString(h[:15])
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
