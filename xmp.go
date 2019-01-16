package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const exiv2 = "./utils/exiv2"

var exiv2Regex = regexp.MustCompile(`(?m:^(\w+)\s+(.*))`)

type xmpSettings struct {
	Orientation int `json:"orientation,omitempty"`

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

func loadXmp(path string) (xmp xmpSettings, err error) {
	log.Printf("exiv2 [-Pnv -gXmp.crs. -gExif.Image. %s]\n", path)
	cmd := exec.Command(exiv2, "-Pnv", "-gExif.Image.", "-gXmp.crs.", path)
	out, err := cmd.Output()
	if err != nil {
		return
	}

	m := make(map[string][]byte)
	for _, s := range exiv2Regex.FindAllSubmatch(out, -1) {
		m[string(s[1])] = bytes.TrimSpace(s[2])
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
	loadInt(&xmp.Temperature, m, "Temperature")
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

func saveXmp(path string, xmp *xmpSettings) (err error) {
	opts := []string{}

	if !strings.HasSuffix(path, ".xmp") && !strings.HasSuffix(path, ".dng") {
		opts = append(opts, "-f", "-eX")
	}
	if xmp != nil {
		opts = append(opts, "-m-")
	}

	if len(opts) > 0 {
		opts = append(opts, path)
		log.Printf("exiv2 %v\n", opts)
		cmd := exec.Command(exiv2, opts...)
		if xmp != nil {
			cmd.Stdin = xmp.buffer()
		}
		_, err = cmd.Output()
	}

	return err
}

// nolint: errcheck
func (xmp *xmpSettings) buffer() *bytes.Buffer {
	buf := bytes.Buffer{}

	// orientation
	fmt.Fprintf(&buf, `
		set Exif.Image.Orientation %[1]d
		set Xmp.tiff.Orientation %[1]d
		set Xmp.crs.ProcessVersion %.1f`,
		xmp.Orientation, xmp.Process)

	// profile
	fmt.Fprintf(&buf, `
		set Xmp.crs.ConvertToGrayscale %t`, xmp.Grayscale)
	if xmp.Profile != "" {
		fmt.Fprintf(&buf, `
			set Xmp.crs.CameraProfile %s`,
			xmp.Profile)
	}

	// white balance
	switch xmp.WhiteBalance {
	case "":
		buf.WriteString(`
			del Xmp.crs.WhiteBalance
			del Xmp.crs.Temperature
			del Xmp.crs.Tint`)
	case "Custom":
		fmt.Fprintf(&buf, `
			set Xmp.crs.WhiteBalance Custom
			set Xmp.crs.Temperature  %d
			set Xmp.crs.Tint         %d`,
			xmp.Temperature, xmp.Tint)
	default:
		fmt.Fprintf(&buf, `
			set Xmp.crs.WhiteBalance %s
			del Xmp.crs.Temperature
			del Xmp.crs.Tint`,
			xmp.WhiteBalance)
	}

	// tone
	if xmp.AutoTone {
		buf.WriteString(`
			set Xmp.crs.AutoTone       True
			set Xmp.crs.AutoExposure   True
			set Xmp.crs.AutoContrast   True
			set Xmp.crs.AutoShadows    True
			set Xmp.crs.AutoBrightness True
			del Xmp.crs.Exposure
			del Xmp.crs.Contrast
			del Xmp.crs.Shadows
			del Xmp.crs.Brightness
			`)
	} else {
		fmt.Fprintf(&buf, `
			del Xmp.crs.AutoTone
			del Xmp.crs.AutoExposure
			del Xmp.crs.AutoContrast
			del Xmp.crs.AutoShadows
			del Xmp.crs.AutoBrightness
			set Xmp.crs.Exposure       %+.2f
			set Xmp.crs.Contrast       %+d
			set Xmp.crs.Shadows        %d
			set Xmp.crs.Brightness     %d
			set Xmp.crs.Exposure2012   %+.2f
			set Xmp.crs.Contrast2012   %+d
			set Xmp.crs.Highlights2012 %+d
			set Xmp.crs.Shadows2012    %+d
			set Xmp.crs.Whites2012     %+d
			set Xmp.crs.Blacks2012     %+d`,
			xmp.oldExposure(), xmp.oldContrast(), xmp.oldShadows(), xmp.oldBrightness(),
			xmp.Exposure, xmp.Contrast, xmp.Highlights, xmp.Shadows, xmp.Whites, xmp.Blacks)
	}

	// presence
	fmt.Fprintf(&buf, `
		set Xmp.crs.Clarity     %+d
		set Xmp.crs.Dehaze      %+d
		set Xmp.crs.Vibrance    %+d
		set Xmp.crs.Saturation  %+d
		set Xmp.crs.Clarity2012 %+d`,
		xmp.oldClarity(), xmp.Dehaze, xmp.Vibrance, xmp.Saturation, xmp.Clarity)

	// detail
	fmt.Fprintf(&buf, `
		set Xmp.crs.Sharpness           %d
		set Xmp.crs.LuminanceSmoothing  %d
		set Xmp.crs.ColorNoiseReduction %d`,
		xmp.Sharpness, xmp.LuminanceNR, xmp.ColorNR)

	// lens corrections
	fmt.Fprintf(&buf, `
		set Xmp.crs.AutoLateralCA     %d
		set Xmp.crs.LensProfileEnable %d`,
		btoi(xmp.AutoLateralCA), btoi(xmp.LensProfile))

	return &buf
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
