package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"rethinkraw/dng"
	"rethinkraw/internal/config"
	"rethinkraw/internal/util"

	"github.com/ncruces/go-exiftool"
)

type xmpSettings struct {
	Filename    string `json:"-"`
	Orientation int    `json:"orientation,omitempty"`

	Process  float32  `json:"process,omitempty"`
	Profile  string   `json:"profile,omitempty"`
	Profiles []string `json:"profiles,omitempty"`

	WhiteBalance string `json:"whiteBalance,omitempty"`
	Temperature  int    `json:"temperature,omitempty"`
	Tint         int    `json:"tint"`

	AutoTone   bool    `json:"autoTone"`
	Exposure   float32 `json:"exposure"`
	Contrast   int     `json:"contrast"`
	Highlights int     `json:"highlights"`
	Shadows    int     `json:"shadows"`
	Whites     int     `json:"whites"`
	Blacks     int     `json:"blacks"`
	Texture    int     `json:"texture"`
	Clarity    int     `json:"clarity"`
	Dehaze     int     `json:"dehaze"`
	Vibrance   int     `json:"vibrance"`
	Saturation int     `json:"saturation"`
	ToneCurve  string  `json:"toneCurve,omitempty"`

	Sharpness   int `json:"sharpness"`
	LuminanceNR int `json:"luminanceNR"`
	ColorNR     int `json:"colorNR"`

	LensProfile   bool `json:"lensProfile"`
	AutoLateralCA bool `json:"autoLateralCA"`
}

type xmpWhiteBalance struct {
	Temperature int `json:"temperature,omitempty"`
	Tint        int `json:"tint"`
}

func loadXMP(path string) (xmp xmpSettings, err error) {
	log.Print("exiftool (load xmp)...")
	out, err := exifserver.Command("--printConv", "-short2", "-fast2",
		"-Orientation", "-Make", "-Model", "-XMP-crs:all", path)
	if err != nil {
		return xmp, err
	}

	m := make(map[string][]byte)
	if err := exiftool.Unmarshal(out, m); err != nil {
		return xmp, err
	}

	// defaults (will be overwritten)
	xmp.Process = 11.0
	xmp.Profile = "Adobe Color"
	xmp.WhiteBalance = "As Shot"
	xmp.ToneCurve = "Linear"
	xmp.Sharpness = 40
	xmp.ColorNR = 25

	// legacy with defaults (will be upgraded/overwritten)
	shadows, brightness, contrast, clarity := 5, 50, 25, 0
	loadString(&xmp.ToneCurve, m, "ToneCurveName")
	loadBool(&xmp.AutoTone, m, "AutoExposure")
	loadFloat32(&xmp.Exposure, m, "Exposure")
	loadInt(&brightness, m, "Brightness")
	loadInt(&contrast, m, "Contrast")
	loadInt(&shadows, m, "Shadows")
	loadInt(&clarity, m, "Clarity")
	xmp.update(shadows, brightness, contrast, clarity)

	// orientation
	loadInt(&xmp.Orientation, m, "Orientation")

	// process/profile
	var process float32
	var grayscale bool
	var profile, look string
	loadFloat32(&process, m, "ProcessVersion")
	loadString(&profile, m, "CameraProfile")
	loadString(&look, m, "LookName")
	loadBool(&grayscale, m, "ConvertToGrayscale")

	adobe, profiles := getCameraProfiles(string(m["Make"]), string(m["Model"]))

	if process != 0 {
		xmp.Process = process
	}
	if process != 0 || profile != "" || look != "" {
		switch {
		case util.Contains(defaultProfiles, look) && (profile == adobe || profile == ""):
			xmp.Profile = look
		case util.Contains(profiles, profile) && look == "":
			xmp.Profile = profile
		case look == "" && (profile == adobe || profile == ""):
			xmp.Profile = "Adobe Standard"
		default:
			xmp.Profile = "Custom"
		}
		if xmp.Profile == "Adobe Standard" && grayscale {
			xmp.Profile = "Adobe Standard B&W"
		}
	}
	xmp.Profiles = profiles

	// curve
	loadString(&xmp.ToneCurve, m, "ToneCurveName2012")
	switch xmp.ToneCurve {
	case "Linear", "Medium Contrast", "Strong Contrast":
	default:
		xmp.ToneCurve = "Custom"
	}

	// white balance
	loadString(&xmp.WhiteBalance, m, "WhiteBalance")
	loadInt(&xmp.Temperature, m, "ColorTemperature")
	loadInt(&xmp.Tint, m, "Tint")

	// tone
	loadBool(&xmp.AutoTone, m, "AutoTone")
	loadFloat32(&xmp.Exposure, m, "Exposure2012")
	loadInt(&xmp.Contrast, m, "Contrast2012")
	loadInt(&xmp.Highlights, m, "Highlights2012")
	loadInt(&xmp.Shadows, m, "Shadows2012")
	loadInt(&xmp.Whites, m, "Whites2012")
	loadInt(&xmp.Blacks, m, "Blacks2012")

	// presence
	loadInt(&xmp.Texture, m, "Texture")
	loadInt(&xmp.Dehaze, m, "Dehaze")
	loadInt(&xmp.Vibrance, m, "Vibrance")
	loadInt(&xmp.Saturation, m, "Saturation")
	loadInt(&xmp.Clarity, m, "Clarity2012")

	// detail
	loadInt(&xmp.Sharpness, m, "Sharpness")
	loadInt(&xmp.LuminanceNR, m, "LuminanceSmoothing")
	loadInt(&xmp.ColorNR, m, "ColorNoiseReduction")

	// lens corrections
	loadBool(&xmp.LensProfile, m, "LensProfileEnable")
	loadBool(&xmp.AutoLateralCA, m, "AutoLateralCA")

	return xmp, nil
}

func editXMP(path string, xmp xmpSettings) error {
	// no process means don't edit
	if xmp.Process == 0 {
		return nil
	}

	// zip means shorter xml output, not compression
	opts := []string{
		"--printConv", "-zip", "-sep", "; ",
		"-XMP-crs:ProcessVersion=" + fmt.Sprintf("%.1f", xmp.Process)}

	// filename
	if xmp.Filename != "" {
		name := filepath.Base(xmp.Filename)
		opts = append(opts, "-XMP-crs:RawFileName="+name)
		if ext := filepath.Ext(xmp.Filename); ext != "" {
			opts = append(opts, "-XMP-photoshop:SidecarForExtension="+ext[1:])
		}
	}

	// orientation
	if xmp.Orientation != 0 {
		opts = append(opts, "-Orientation="+strconv.Itoa(xmp.Orientation))
	}
	// profile
	if xmp.Profile != "" && xmp.Profile != "Custom" {
		if settings, ok := profileSettings[xmp.Profile]; ok {
			opts = append(opts, settings...)
		} else {
			opts = append(opts,
				"-XMP-crs:CameraProfile="+xmp.Profile,
				"-XMP-crs:ConvertToGrayscale=",
				"-XMP-crs:Look*=")
		}
	}

	// white balance
	if xmp.WhiteBalance == "Custom" {
		opts = append(opts,
			"-XMP-crs:ColorTemperature="+strconv.Itoa(xmp.Temperature),
			"-XMP-crs:Tint="+strconv.Itoa(xmp.Tint),
			"-XMP-crs:WhiteBalance=Custom")
	} else if xmp.WhiteBalance != "" {
		opts = append(opts,
			"-XMP-crs:WhiteBalance="+xmp.WhiteBalance,
			"-XMP-crs:ColorTemperature=",
			"-XMP-crs:Tint=")
	}

	// tone
	if xmp.AutoTone {
		opts = append(opts,
			"-XMP-crs:AutoTone=true",
			"-XMP-crs:AutoExposure=true",
			"-XMP-crs:AutoContrast=true",
			"-XMP-crs:AutoShadows=true",
			"-XMP-crs:AutoBrightness=true",
			"-XMP-crs:Exposure=",
			"-XMP-crs:Contrast=",
			"-XMP-crs:Shadows=",
			"-XMP-crs:Brightness=",
			"-XMP-crs:Exposure2012=",
			"-XMP-crs:Contrast2012=",
			"-XMP-crs:Highlights2012=",
			"-XMP-crs:Shadows2012=",
			"-XMP-crs:Whites2012=",
			"-XMP-crs:Blacks2012=",
			"-XMP-crs:Vibrance=",
			"-XMP-crs:Saturation=")
	} else {
		opts = append(opts,
			"-XMP-crs:AutoTone=",
			"-XMP-crs:AutoExposure=",
			"-XMP-crs:AutoContrast=",
			"-XMP-crs:AutoShadows=",
			"-XMP-crs:AutoBrightness=",
			"-XMP-crs:Exposure="+fmt.Sprintf("%.2f", xmp.oldExposure()),
			"-XMP-crs:Contrast="+strconv.Itoa(xmp.oldContrast()),
			"-XMP-crs:Shadows="+strconv.Itoa(xmp.oldShadows()),
			"-XMP-crs:Brightness="+strconv.Itoa(xmp.oldBrightness()),
			"-XMP-crs:Exposure2012="+fmt.Sprintf("%.2f", xmp.Exposure),
			"-XMP-crs:Contrast2012="+strconv.Itoa(xmp.Contrast),
			"-XMP-crs:Highlights2012="+strconv.Itoa(xmp.Highlights),
			"-XMP-crs:Shadows2012="+strconv.Itoa(xmp.Shadows),
			"-XMP-crs:Whites2012="+strconv.Itoa(xmp.Whites),
			"-XMP-crs:Blacks2012="+strconv.Itoa(xmp.Blacks),
			"-XMP-crs:Vibrance="+strconv.Itoa(xmp.Vibrance),
			"-XMP-crs:Saturation="+strconv.Itoa(xmp.Saturation))
	}

	switch xmp.ToneCurve {
	case "Linear":
		opts = append(opts, "-XMP-crs:ToneCurve*=")
	case "Medium Contrast":
		opts = append(opts,
			"-XMP-crs:ToneCurve*=",
			"-XMP-crs:ToneCurveName=Medium Contrast",
			"-XMP-crs:ToneCurveName2012=Medium Contrast",
			"-XMP-crs:ToneCurve=0, 0; 32, 22; 64, 56; 128, 128; 192, 196; 255, 255",
			"-XMP-crs:ToneCurvePV2012=0, 0; 32, 22; 64, 56; 128, 128; 192, 196; 255, 255",
		)
	case "Strong Contrast":
		opts = append(opts,
			"-XMP-crs:ToneCurve*=",
			"-XMP-crs:ToneCurveName=Strong Contrast",
			"-XMP-crs:ToneCurveName2012=Strong Contrast",
			"-XMP-crs:ToneCurve=0, 0; 32, 16; 64, 50; 128, 128; 192, 202; 255, 255",
			"-XMP-crs:ToneCurvePV2012=0, 0; 32, 16; 64, 50; 128, 128; 192, 202; 255, 255",
		)
	}

	// presence
	opts = append(opts,
		"-XMP-crs:Clarity="+strconv.Itoa(xmp.oldClarity()),
		"-XMP-crs:Texture="+strconv.Itoa(xmp.Texture),
		"-XMP-crs:Dehaze="+strconv.Itoa(xmp.Dehaze),
		"-XMP-crs:Clarity2012="+strconv.Itoa(xmp.Clarity))

	// detail
	opts = append(opts,
		"-XMP-crs:Sharpness="+strconv.Itoa(xmp.Sharpness),
		"-XMP-crs:LuminanceSmoothing="+strconv.Itoa(xmp.LuminanceNR),
		"-XMP-crs:ColorNoiseReduction="+strconv.Itoa(xmp.ColorNR))

	// lens corrections
	opts = append(opts,
		"-XMP-crs:AutoLateralCA="+strconv.Itoa(util.Btoi(xmp.AutoLateralCA)),
		"-XMP-crs:LensProfileEnable="+strconv.Itoa(util.Btoi(xmp.LensProfile)))

	opts = append(opts, "-overwrite_original", path)

	log.Print("exiftool (edit xmp)...")
	_, err := exifserver.Command(opts...)
	return err
}

func extractXMP(path, dest string) error {
	log.Print("exiftool (extract xmp)...")
	_, err := exifserver.Command("--printConv", "-fast2",
		"-tagsFromFile", path, "-scanForXMP",
		"-Orientation", "-Make", "-Model", "-all:all",
		"-overwrite_original", dest)
	return err
}

func computeWhiteBalance(meta, pixels string, coords []float64) (wb xmpWhiteBalance, err error) {
	log.Print("exiftool (load camera profile)...")

	out, err := exifserver.Command("--printConv", "-short2", "-fast2",
		"-EXIF:CalibrationIlluminant?", "-EXIF:ColorMatrix?",
		"-EXIF:CameraCalibration?", "-EXIF:AnalogBalance",
		"-EXIF:AsShotNeutral", "-EXIF:AsShotWhiteXY", meta)
	if err != nil {
		return wb, err
	}

	m := make(map[string][]byte)
	if err := exiftool.Unmarshal(out, m); err != nil {
		return wb, err
	}
	// ColorMatrix1 is required for all non-monochrome DNGs.
	if _, ok := m["ColorMatrix1"]; !ok {
		return wb, errors.New("unsupported monochrome camera")
	}

	var profile dng.CameraProfile
	var neutral, whiteXY []float64
	var illuminant1, illuminant2 int
	loadFloat64s(&profile.ColorMatrix1, m, "ColorMatrix1")
	loadFloat64s(&profile.ColorMatrix2, m, "ColorMatrix2")
	loadFloat64s(&profile.CameraCalibration1, m, "CameraCalibration1")
	loadFloat64s(&profile.CameraCalibration2, m, "CameraCalibration2")
	loadFloat64s(&profile.AnalogBalance, m, "AnalogBalance")
	loadFloat64s(&neutral, m, "AsShotNeutral")
	loadFloat64s(&whiteXY, m, "AsShotWhiteXY")
	loadInt(&illuminant1, m, "CalibrationIlluminant1")
	loadInt(&illuminant2, m, "CalibrationIlluminant2")
	profile.CalibrationIlluminant1 = dng.LightSource(illuminant1)
	profile.CalibrationIlluminant2 = dng.LightSource(illuminant2)

	if len(coords) != 2 {
		switch {
		case len(whiteXY) == 2:
			wb.Temperature, wb.Tint = dng.GetTemperatureFromXY(whiteXY[0], whiteXY[1])
		case len(neutral) >= 2:
			wb.Temperature, wb.Tint, err = profile.GetTemperature(neutral)
		}
		return wb, err
	}

	if len(profile.ColorMatrix1) != 3*3 {
		return wb, errors.New("unsupported 4-color camera")
	}

	neutral, err = getMultipliers(pixels, coords)
	if err != nil {
		return wb, err
	}

	wb.Temperature, wb.Tint, err = profile.GetTemperature(neutral)
	return wb, err
}

func (xmp *xmpSettings) update(shadows, brightness, contrast, clarity int) {
	xmp.Exposure += float32(brightness-50) / 50
	xmp.Contrast = 100 * (contrast - 25) / 75

	if shadows <= 5 {
		xmp.Blacks = 5 * (5 - shadows)
	} else {
		xmp.Blacks = 25 * (5 - shadows) / 95
	}

	if clarity > 0 {
		xmp.Clarity = clarity / 2
	} else {
		xmp.Clarity = clarity
	}
}

func (xmp *xmpSettings) oldExposure() float32 {
	if xmp.Exposure > +4 {
		return +4
	}
	if xmp.Exposure < -4 {
		return -4
	}
	return xmp.Exposure
}

func (xmp *xmpSettings) oldContrast() int {
	return 25 + 75*xmp.Contrast/100
}

func (xmp *xmpSettings) oldShadows() int {
	if xmp.Blacks >= +25 {
		return 0
	}
	if xmp.Blacks >= 0 {
		return 5 - xmp.Blacks/5
	}
	if xmp.Blacks >= -25 {
		return 5 - 95*xmp.Blacks/25
	}
	return 100
}

func (xmp *xmpSettings) oldBrightness() int {
	if xmp.Exposure > +6 {
		return 150
	}
	if xmp.Exposure > +4 {
		return int(50 + 50*(xmp.Exposure-4))
	}
	if xmp.Exposure < -5 {
		return 0
	}
	if xmp.Exposure < -4 {
		return int(50 + 50*(xmp.Exposure+4))
	}
	return 50
}

func (xmp *xmpSettings) oldClarity() int {
	if xmp.Clarity >= 50 {
		return 100
	}
	if xmp.Clarity > 0 {
		return 2 * xmp.Clarity
	}
	return xmp.Clarity
}

func loadString(dst *string, m map[string][]byte, key string) {
	if v, ok := m[key]; ok {
		*dst = string(v)
	}
}

func loadInt(dst *int, m map[string][]byte, key string) {
	if v, ok := m[key]; ok {
		i, err := strconv.Atoi(string(v))
		if err == nil {
			*dst = i
		}
	}
}

func loadBool(dst *bool, m map[string][]byte, key string) {
	if v, ok := m[key]; ok {
		b, err := strconv.ParseBool(string(v))
		if err == nil {
			*dst = b
		}
	}
}

func loadFloat32(dst *float32, m map[string][]byte, key string) {
	if v, ok := m[key]; ok {
		f, err := strconv.ParseFloat(string(v), 32)
		if err == nil {
			*dst = float32(f)
		}
	}
}

func loadFloat64s(dst *[]float64, m map[string][]byte, key string) {
	if v, ok := m[key]; ok {
		var fs []float64
		for _, s := range bytes.Split(v, []byte(" ")) {
			f, err := strconv.ParseFloat(string(s), 64)
			if err != nil {
				return
			}
			fs = append(fs, f)
		}
		*dst = fs
	}
}

var dcrawThumbRegex = regexp.MustCompile(`Thumb size: +(\d+) x (\d+)`)

func dngPreview(path string) string {
	log.Print("dcraw (get thumb size)...")
	cmd := exec.Command(config.Dcraw, "-i", "-v", path)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	var max int
	if match := dcrawThumbRegex.FindSubmatch(out); match != nil {
		width, _ := strconv.Atoi(string(match[1]))
		height, _ := strconv.Atoi(string(match[2]))
		if width > height {
			max = width
		} else {
			max = height
		}
	}

	switch {
	case max > 1024:
		return "p2"
	case max > 256:
		return "p1"
	default:
		return "p0"
	}
}

func getRawPixels(path string) error {
	log.Print("dcraw (get raw pixels)...")
	cmd := exec.Command(config.DcrawEmu,
		"-r", "1", "1", "1", "1",
		"-o", "0",
		"-h",
		"-4",
		"-t", "0",
		"-Z", "ppm",
		path)
	err := cmd.Run()
	if err == nil {
		return nil
	}
	cmd = exec.Command(config.Dcraw,
		"-r", "1", "1", "1", "1",
		"-o", "0",
		"-h",
		"-4",
		"-t", "0",
		path)
	_, err = cmd.Output()
	return err
}

func getMultipliers(path string, coords []float64) ([]float64, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var format, width, height int
	n, _ := fmt.Fscanf(bytes.NewReader(data), "P%d\n%d %d\n65535\n", &format, &width, &height)
	if n == 3 {
		for i := 0; i < 3; i++ {
			data = data[bytes.IndexByte(data, '\n')+1:]
		}

		if format == 6 && len(data) == 6*width*height {
			x := int(coords[0]*float64(width)) - 1
			y := int(coords[1]*float64(height)) - 1
			if x < 0 {
				x = 0
			}
			if y < 0 {
				y = 0
			}
			if x > width-4 {
				x = width - 4
			}
			if y >= height-4 {
				y = height - 4
			}

			var r, g, b int
			for yy := 0; yy < 4; yy++ {
				for xx := 0; xx < 4; xx++ {
					i := (x+xx)*6 + (y+yy)*6*width
					r += 256*int(data[0+i]) + int(data[1+i])
					g += 256*int(data[2+i]) + int(data[3+i])
					b += 256*int(data[4+i]) + int(data[5+i])
				}
			}

			if r == g && b == g {
				return nil, errors.New("unsupported camera")
			}

			var multipliers [3]float64
			multipliers[0] = float64(r) / float64(g)
			multipliers[1] = float64(g) / float64(g)
			multipliers[2] = float64(b) / float64(g)
			return multipliers[:], nil
		}
	}

	return nil, errors.New("unsupported pixel map")
}
