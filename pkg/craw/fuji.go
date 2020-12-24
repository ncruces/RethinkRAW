package craw

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func FujifilmCameraProfiles(model string) ([]string, error) {
	f, err := os.Open(EmbedProfiles)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("_kProfileData_Fujifilm_Full_%s_Camera_", strings.ReplaceAll(model, " ", "_"))

	scan := bufio.NewScanner(f)
	scan.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if begin := bytes.Index(data, []byte(prefix)); begin >= 0 {
			if end := bytes.IndexByte(data[begin+len(prefix):], 0); end >= 0 {
				last := begin + len(prefix) + end
				return last + 1, data[begin+len(prefix) : last], nil
			}
			advance = begin
		} else {
			advance = len(data) - len(prefix) + 1
		}

		if atEOF {
			return 0, nil, io.EOF
		}
		return advance, nil, nil
	})

	var profiles []string
	for scan.Scan() {
		var name string
		id := strings.ToUpper(scan.Text())

		switch strings.TrimSuffix(id, "_V2") {
		case "PROVIA_STANDARD":
			name = "Camera PROVIA/Standard"
		case "VELVIA_VIVID":
			name = "Camera Velvia/Vivid"
		case "ASTIA_SOFT":
			name = "Camera ASTIA/Soft"
		case "PRO_NEG_HI":
			name = "Camera Pro Neg Hi"
		case "PRO_NEG_STD":
			name = "Camera Pro Neg Std"
		case "MONOCHROME":
			name = "Camera Monochrome"
		case "MONOCHROME_YE_FILTER":
			name = "Camera Monochrome+Ye Filter"
		case "MONOCHROME_R_FILTER":
			name = "Camera Monochrome+R Filter"
		case "MONOCHROME_G_FILTER":
			name = "Camera Monochrome+G Filter"
		case "ACROS":
			name = "Camera ACROS"
		case "ACROS_YE_FILTER":
			name = "Camera ACROS+Ye Filter"
		case "ACROS_R_FILTER":
			name = "Camera ACROS+R Filter"
		case "ACROS_G_FILTER":
			name = "Camera ACROS+G Filter"
		case "CLASSIC_CHROME":
			name = "Camera CLASSIC CHROME"
		case "ETERNA_CINEMA":
			name = "Camera ETERNA/Cinema"

		default:
			continue
		}
		if strings.HasSuffix(id, "_V2") {
			name += " v2"
		}

		if !contains(profiles, name) {
			profiles = append(profiles, name)
		}
	}

	return profiles, scan.Err()
}

func contains(a []string, s string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}
