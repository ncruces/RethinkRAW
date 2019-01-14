package main

import (
	"image"
	"io"
	"math"
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

	return previewJPEG(wk.Edit())
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

	data, err = exportJPEG(wk.Temp(), exp)
	if err != nil {
		return
	}

	err = os.Remove(wk.Temp())
	return
}

type exportSettings struct {
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
	hash := md5sum(filepath.ToSlash(path))

	wk.base = filepath.Join(tempDir, hash) + string(filepath.Separator)
	wk.ext = filepath.Ext(path)

	err = os.MkdirAll(wk.base, 0700)
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
