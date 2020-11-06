package main

import (
	"bufio"
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

	Sharpness   int `json:"sharpness"`
	LuminanceNR int `json:"luminanceNR"`
	ColorNR     int `json:"colorNR"`

	LensProfile   bool `json:"lensProfile"`
	AutoLateralCA bool `json:"autoLateralCA"`
}

type xmpWhiteBalance struct {
	Temperature int `json:"temperature"`
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

	// legacy with defaults (will be upgraded/overwritten)
	shadows, brightness, contrast, clarity := 5, 50, 25, 0
	loadBool(&xmp.AutoTone, m, "AutoExposure")
	loadFloat32(&xmp.Exposure, m, "Exposure")
	loadInt(&brightness, m, "Brightness")
	loadInt(&contrast, m, "Contrast")
	loadInt(&shadows, m, "Shadows")
	loadInt(&clarity, m, "Clarity")
	xmp.update(shadows, brightness, contrast, clarity)

	// defaults (will be overwritten)
	xmp.Process = 11.0
	xmp.Profile = "Adobe Color"
	xmp.WhiteBalance = "As Shot"
	xmp.Sharpness = 40
	xmp.ColorNR = 25

	// orientation
	loadInt(&xmp.Orientation, m, "Orientation")

	// process/profile
	var grayscale bool
	loadFloat32(&xmp.Process, m, "ProcessVersion")
	loadString(&xmp.Profile, m, "CameraProfile")
	loadBool(&grayscale, m, "ConvertToGrayscale")
	if grayscale && xmp.Profile == "Adobe Standard" {
		xmp.Profile = "Adobe Standard B&W"
	}

	// load camera profiles
	xmp.Profiles = append(profiles, getCameraProfiles(string(m["Make"]), string(m["Model"]))...)
	if index(xmp.Profiles, xmp.Profile) < 0 {
		xmp.Profiles = append(xmp.Profiles, xmp.Profile)
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

func editXMP(path string, xmp *xmpSettings) error {
	// zip means shorter xml output, not compression
	opts := []string{"--printConv", "-zip"}

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
	// process
	if xmp.Process != 0 {
		opts = append(opts, "-XMP-crs:ProcessVersion="+fmt.Sprintf("%.1f", xmp.Process))
	}
	// profile
	if xmp.Profile != "" {
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
	opts = append(opts, "-XMP-crs:WhiteBalance="+xmp.WhiteBalance)
	if xmp.WhiteBalance == "Custom" {
		opts = append(opts,
			"-XMP-crs:ColorTemperature="+strconv.Itoa(xmp.Temperature),
			"-XMP-crs:Tint="+strconv.Itoa(xmp.Tint))
	} else {
		opts = append(opts,
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
		"-XMP-crs:AutoLateralCA="+strconv.Itoa(btoi(xmp.AutoLateralCA)),
		"-XMP-crs:LensProfileEnable="+strconv.Itoa(btoi(xmp.LensProfile)))

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

func extractWhiteBalance(path string, coords []int) (map[string]xmpWhiteBalance, error) {
	log.Print("exiftool (load camera profile)...")

	out, err := exifserver.Command("--printConv", "-short2", "-fast2",
		"-EXIF:CalibrationIlluminant?", "-EXIF:ColorMatrix?",
		"-EXIF:CameraCalibration?", "-EXIF:AnalogBalance",
		"-EXIF:AsShotNeutral", "-EXIF:AsShotWhiteXY", path)
	if err != nil {
		return nil, err
	}

	m := make(map[string][]byte)
	if err := exiftool.Unmarshal(out, m); err != nil {
		return nil, err
	}
	// ColorMatrix1 is required for all non-monochrome DNGs.
	if _, ok := m["ColorMatrix1"]; !ok {
		return nil, nil
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

	res := make(map[string]xmpWhiteBalance)

	switch {
	case len(whiteXY) == 2:
		var wb xmpWhiteBalance
		wb.Temperature, wb.Tint = dng.GetTemperature(whiteXY[0], whiteXY[1])
		res["As Shot"] = wb

	case len(neutral) >= 2:
		var wb xmpWhiteBalance
		wb.Temperature, wb.Tint, err = profile.GetTemperature(neutral)
		if err != nil {
			return nil, err
		} else {
			res["As Shot"] = wb
		}
	}

	if len(coords) != 2 {
		return res, nil
	}

	mul, err := dngMultipliers(path, "-A",
		strconv.Itoa(int(coords[0])),
		strconv.Itoa(int(coords[1])),
		"2", "2")
	if err != nil {
		return res, err
	}
	if len(mul) > 1 {
		var wb xmpWhiteBalance
		mul := mul[:len(profile.ColorMatrix1)/3] // TODO: are we sure?
		wb.Temperature, wb.Tint, err = profile.GetTemperature(mul)
		if err != nil {
			return nil, err
		} else {
			res["Custom"] = wb
		}
	}

	return res, err
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
	cmd := exec.Command(dcraw, "-i", "-v", path)
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

func dngMultipliers(path string, arg ...string) ([]float64, error) {
	log.Print("dcraw (get multipliers)...")
	cmd := exec.Command(dcraw, append(arg, "-v", "-d", "-c", path)...)
	cmd.Stdout = ioutil.Discard

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	var msg bytes.Buffer
	var multipliers []float64
	for r := bufio.NewReader(stderr); ; {
		buf, err := r.ReadBytes('\n')
		msg.Write(buf)
		if err != nil {
			break
		}
		if bytes.HasPrefix(buf, []byte("multipliers ")) {
			multipliers = make([]float64, 4)
			data := bytes.SplitN(buf[:len(buf)-1], []byte(" "), 5)[1:]
			multipliers[0], _ = strconv.ParseFloat(string(data[0]), 64)
			multipliers[1], _ = strconv.ParseFloat(string(data[1]), 64)
			multipliers[2], _ = strconv.ParseFloat(string(data[2]), 64)
			multipliers[3], _ = strconv.ParseFloat(string(data[3]), 64)
			break
		}
	}
	stderr.Close()

	err = cmd.Wait()
	var eerr *exec.ExitError
	if errors.As(err, &eerr) {
		eerr.Stderr = msg.Bytes()
	}

	if multipliers != nil {
		multipliers[0] = 1 / multipliers[0]
		multipliers[1] = 1 / multipliers[1]
		multipliers[2] = 1 / multipliers[2]
		multipliers[3] = 1 / multipliers[3]
		return multipliers, nil
	}
	return nil, err
}
