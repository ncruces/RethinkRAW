package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
)

var dcrawThumbRegex = regexp.MustCompile(`Thumb size: +(\d+) x (\d+)`)
var exiftoolRegex = regexp.MustCompile(`(?m:^(\w+): (.*))`)

type xmpSettings struct {
	Filename    string `json:"-"`
	Orientation int    `json:"orientation,omitempty"`

	Process   float32 `json:"process,omitempty"`
	Profile   string  `json:"profile,omitempty"`
	Grayscale bool    `json:"grayscale"`

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

func loadXMP(path string) (xmp xmpSettings, err error) {
	log.Print("exiftool (load xmp)...")
	out, err := exifserver.Command("-S", "-n", "-fast2", "-orientation", "-xmp-crs:*", path)
	if err != nil {
		return
	}

	m := make(map[string][]byte)
	for _, s := range exiftoolRegex.FindAllSubmatch(out, -1) {
		m[string(s[1])] = bytes.TrimRight(s[2], "\r")
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
	xmp.Profile = "Adobe Standard"
	xmp.WhiteBalance = "As Shot"
	xmp.Sharpness = 40
	xmp.ColorNR = 25
	xmp.LensProfile = true
	xmp.AutoLateralCA = true

	// orientation
	loadInt(&xmp.Orientation, m, "Orientation")

	// process/profile
	loadFloat32(&xmp.Process, m, "ProcessVersion")
	loadString(&xmp.Profile, m, "CameraProfile")
	loadBool(&xmp.Grayscale, m, "ConvertToGrayscale")

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

	return
}

func saveXMP(path string, xmp *xmpSettings) (err error) {
	opts := []string{"-n", "-z"}

	// filename
	if xmp.Filename != "" {
		name := filepath.Base(xmp.Filename)
		ext := filepath.Ext(xmp.Filename)
		opts = append(opts,
			"-XMP-crs:RawFileName="+name,
			"-XMP-photoshop:SidecarForExtension="+ext)
	}

	// orientation, process, grayscale
	opts = append(opts,
		"-Orientation="+strconv.Itoa(xmp.Orientation),
		"-XMP-crs:ProcessVersion="+fmt.Sprintf("%.1f", xmp.Process),
		"-XMP-crs:ConvertToGrayscale="+strconv.FormatBool(xmp.Grayscale))

	// profile
	if xmp.Profile != "" {
		opts = append(opts, "-XMP-crs:CameraProfile="+xmp.Profile)
	}

	// white balance
	opts = append(opts, "-XMP-crs:WhiteBalance="+xmp.WhiteBalance)
	switch xmp.WhiteBalance {
	case "As Shot", "Auto":
	case "Custom":
		opts = append(opts,
			"-XMP-crs:ColorTemperature="+strconv.Itoa(xmp.Temperature),
			"-XMP-crs:Tint="+strconv.Itoa(xmp.Tint))
	default:
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
			"-XMP-crs:Brightness=")
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
			"-XMP-crs:Blacks2012="+strconv.Itoa(xmp.Blacks))
	}

	// presence
	opts = append(opts,
		"-XMP-crs:Clarity="+strconv.Itoa(xmp.oldClarity()),
		"-XMP-crs:Dehaze="+strconv.Itoa(xmp.Dehaze),
		"-XMP-crs:Vibrance="+strconv.Itoa(xmp.Vibrance),
		"-XMP-crs:Saturation="+strconv.Itoa(xmp.Saturation),
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

	log.Print("exiftool (save xmp)...")
	_, err = exifserver.Command(opts...)
	return
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

func dngPreview(path string) string {
	log.Print("dcraw thumb-size ...")
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
