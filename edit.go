package main

import (
	"bytes"
	"image"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/rethinkraw/pkg/osutil"
	"github.com/ncruces/rethinkraw/pkg/xmp"
)

func loadEdit(path string) (xmp xmpSettings, err error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return xmp, err
	}
	defer wk.close()

	return loadXMP(wk.origXMP())
}

func saveEdit(path string, xmp xmpSettings) error {
	wk, err := openWorkspace(path)
	if err != nil {
		return err
	}
	defer wk.close()

	if xmp.WhiteBalance == "Camera Matching…" {
		xmp.WhiteBalance = cameraMatchingWhiteBalance(wk.orig())
	}

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

		err = osutil.Move(wk.temp(), dest+".bak")
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

	if xmp.WhiteBalance == "Camera Matching…" {
		xmp.WhiteBalance = cameraMatchingWhiteBalance(wk.orig())
	}

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
		// create edit.dng (downscaled to 2560 on the widest side)

		err = editXMP(wk.origXMP(), xmp)
		if err != nil {
			return nil, err
		}

		err = runDNGConverter(wk.orig(), wk.edit(), 2560, nil)
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

	if xmp.WhiteBalance == "Camera Matching…" {
		xmp.WhiteBalance = cameraMatchingWhiteBalance(wk.orig())
	}

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

		return os.ReadFile(wk.temp())
	} else {
		data, err := exportJPEG(wk.temp(), exp)
		if err != nil || exp.Resample {
			return data, err
		}

		go func() {
			writer.Write(data)
			writer.Close()
		}()
		return io.ReadAll(reader)
	}
}

func exportPath(path string, exp exportSettings) string {
	var ext string
	if exp.DNG {
		ext = ".dng"
	} else {
		ext = ".jpg"
	}
	return strings.TrimSuffix(path, filepath.Ext(path)) + ext
}

func loadWhiteBalance(path string, coords []float64) (wb xmpWhiteBalance, err error) {
	wk, err := openWorkspace(path)
	if err != nil {
		return wb, err
	}
	defer wk.close()

	if !wk.hasEdit {
		// create edit.dng (downscaled to at most 2560 on the widest side)

		err = runDNGConverter(wk.orig(), wk.edit(), 2560, nil)
		if err != nil {
			return wb, err
		}
	}

	if !wk.hasPixels && len(coords) == 2 {
		err = getRawPixels(wk.edit(), wk.pixels())
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
			fit.X = math.MaxInt
			fit.Y = int(mul * float64(size.Y))
		} else {
			fit.X = int(mul * float64(size.X))
			fit.Y = math.MaxInt
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
			if x == 0.0 {
				return math.MaxInt
			}
			i := int(x + 0.5)
			if i < 16 {
				return 16
			} else {
				return i
			}
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
				fit.X = round(mul * short)
				fit.Y = round(mul * long)
			}
		} else {
			fit.X = round(mul * ex.Width)
			fit.Y = round(mul * ex.Height)
		}
	}
	return fit
}

func loadSidecar(src, dst string) error {
	var data []byte
	err := os.ErrNotExist
	ext := filepath.Ext(src)

	if ext != "" {
		// if NAME.xmp is there for NAME.EXT, use it
		name := strings.TrimSuffix(src, ext) + ".xmp"
		data, err = os.ReadFile(name)
		if err == nil && !xmp.IsSidecarForExt(bytes.NewReader(data), ext) {
			err = os.ErrNotExist
		}
	}
	if os.IsNotExist(err) {
		// if NAME.EXT.xmp is there for NAME.EXT, use it
		data, err = os.ReadFile(src + ".xmp")
		if err == nil && !xmp.IsSidecarForExt(bytes.NewReader(data), ext) {
			err = os.ErrNotExist
		}
	}
	if err == nil {
		// copy xmp file
		err = os.WriteFile(dst, data, 0600)
	}
	if err == nil || os.IsNotExist(err) {
		// extract embed XMP data
		return extractXMP(src, dst)
	}
	return err
}

func destSidecar(src string) (string, error) {
	ext := filepath.Ext(src)

	if ext != "" {
		// if NAME.xmp is there for NAME.EXT, use it
		name := strings.TrimSuffix(src, ext) + ".xmp"
		data, err := os.ReadFile(name)
		if err == nil && xmp.IsSidecarForExt(bytes.NewReader(data), ext) {
			return name, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
	}

	// if NAME.EXT.xmp exists, use it
	if _, err := os.Stat(src + ".xmp"); err == nil {
		return src + ".xmp", nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	// if NAME.DNG was edited, use it
	if strings.EqualFold(ext, ".dng") && dngHasEdits(src) {
		return src, nil
	}

	// fallback to NAME.xmp
	return strings.TrimSuffix(src, ext) + ".xmp", nil
}
