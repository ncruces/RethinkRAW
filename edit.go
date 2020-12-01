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

	"rethinkraw/internal/util"
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

func saveEdit(path string, xmp xmpSettings) error {
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

func previewEdit(path string, size int, xmp xmpSettings) ([]byte, error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return nil, err
	}
	defer wk.close()

	if size == 0 {
		// use the original RAW file for a full resolution preview

		err = editXMP(wk.origXMP(), xmp)
		if err != nil {
			return nil, err
		}

		err = runDNGConverter(wk.orig(), wk.temp(), 0, &exportSettings{})
		if err != nil {
			return nil, err
		}

		return previewJPEG(wk.temp())
	} else if wk.hasEdit {
		// use edit.dng (downscaled to at most 2560 on the widest side)

		err = editXMP(wk.edit(), xmp)
		if err != nil {
			return nil, err
		}

		err = runDNGConverter(wk.edit(), wk.temp(), size, nil)
		if err != nil {
			return nil, err
		}

		return previewJPEG(wk.temp())
	} else {
		// create edit.dng (downscaled to at most 2560 on the widest side)

		err = editXMP(wk.origXMP(), xmp)
		if err != nil {
			return nil, err
		}

		err = runDNGConverter(wk.orig(), wk.temp(), 2560, nil)
		if err != nil {
			return nil, err
		}

		err = os.Rename(wk.temp(), wk.edit())
		if err != nil {
			return nil, err
		}

		return previewJPEG(wk.edit())
	}
}

func exportEdit(path string, xmp xmpSettings, exp exportSettings) ([]byte, error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return nil, err
	}
	defer wk.close()

	err = editXMP(wk.origXMP(), xmp)
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

	err = runDNGConverter(wk.orig(), wk.temp(), 0, &exp)
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

func exportName(path string, exp exportSettings) string {
	var ext string
	if exp.DNG {
		ext = ".dng"
	} else {
		ext = ".jpg"
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) + ext
}

func loadWhiteBalance(path string, coords []float64) (wb xmpWhiteBalance, err error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return wb, err
	}
	defer wk.close()

	if !wk.hasEdit {
		// create edit.dng (downscaled to at most 2560 on the widest side)

		err = runDNGConverter(wk.orig(), wk.temp(), 2560, nil)
		if err != nil {
			return wb, err
		}

		err = os.Rename(wk.temp(), wk.edit())
		if err != nil {
			return wb, err
		}
	}

	if len(coords) == 2 && !wk.hasPixels {
		err = getRawPixels(wk.orig())
		if err != nil {
			return wb, err
		}
	}

	return computeWhiteBalance(wk.edit(), wk.pixels(), coords)
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
			fit.X = util.MaxInt
			fit.Y = int(mul * float64(size.Y))
		} else {
			fit.X = int(mul * float64(size.X))
			fit.Y = util.MaxInt
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
			return util.MaxInt
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

	// fallback to NAME.XMP
	return strings.TrimSuffix(src, ext) + ".xmp", nil
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
			return err == io.EOF // assume yes
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
