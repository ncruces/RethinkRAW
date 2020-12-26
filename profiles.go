package main

import (
	"sort"
	"strings"
	"sync"

	"github.com/ncruces/rethinkraw/internal/config"
	"github.com/ncruces/rethinkraw/internal/util"
	"github.com/ncruces/rethinkraw/pkg/craw"
)

var defaultProfiles = []string{
	"Adobe Color", "Adobe Monochrome", "Adobe Landscape", "Adobe Neutral",
	"Adobe Portrait", "Adobe Vivid", "Adobe Standard", "Adobe Standard B&W",
}

var profileSettings = map[string][]string{
	"Adobe Standard": {
		"-XMP-crs:ConvertToGrayscale=",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
	},
	"Adobe Standard B&W": {
		"-XMP-crs:ConvertToGrayscale=True",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
	},
	"Adobe Color": {
		"-XMP-crs:ConvertToGrayscale=",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Color",
		"-XMP-crs:LookUUID=B952C231111CD8E0ECCF14B86BAA7077",
		"-XMP-crs:LookParametersToneCurvePV2012=0, 0; 22, 16; 40, 35; 127, 127; 224, 230; 240, 246; 255, 255",
		"-XMP-crs:LookParametersLookTable=E1095149FDB39D7A057BAB208837E2E1",
	},
	"Adobe Monochrome": {
		"-XMP-crs:ConvertToGrayscale=True",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Monochrome",
		"-XMP-crs:LookUUID=0CFE8F8AB5F63B2A73CE0B0077D20817",
		"-XMP-crs:LookParametersConvertToGrayscale=True",
		"-XMP-crs:LookParametersClarity2012=+8",
		"-XMP-crs:LookParametersToneCurvePV2012=0, 0; 64, 56; 128, 128; 192, 197; 255, 255",
		"-XMP-crs:LookParametersLookTable=73ED6C18DDE909DD7EA2D771F5AC282D",
	},
	"Adobe Landscape": {
		"-XMP-crs:ConvertToGrayscale=",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Landscape",
		"-XMP-crs:LookUUID=6F9C877E84273F4E8271E6B91BEB36A1",
		"-XMP-crs:LookParametersHighlights2012=-12",
		"-XMP-crs:LookParametersShadows2012=+12",
		"-XMP-crs:LookParametersClarity2012=+10",
		"-XMP-crs:LookParametersToneCurvePV2012=0, 0; 64, 60; 128, 128; 192, 196; 255, 255",
		"-XMP-crs:LookParametersLookTable=0B3BFB5CFB7DBF7FF175E98F24D316B0",
	},
	"Adobe Neutral": {
		"-XMP-crs:ConvertToGrayscale=",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Neutral",
		"-XMP-crs:LookUUID=1E8E067A11CD44394A3C36A327BB34D1",
		"-XMP-crs:LookParametersToneCurvePV2012=0, 0; 16, 24; 64, 72; 128, 128; 192, 176; 244, 234; 255, 255",
		"-XMP-crs:LookParametersLookTable=7740BB918B2F6D93D7B95A4DBB78DB23",
	},
	"Adobe Portrait": {
		"-XMP-crs:ConvertToGrayscale=",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Portrait",
		"-XMP-crs:LookUUID=D6496412E06A83789C499DF9540AA616",
		"-XMP-crs:LookParametersToneCurvePV2012=0, 0; 66, 64; 190, 192; 255, 255",
		"-XMP-crs:LookParametersLookTable=E5A76DBB8B3F132A04C01AF45DC2EF1B",
	},
	"Adobe Vivid": {
		"-XMP-crs:ConvertToGrayscale=",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Vivid",
		"-XMP-crs:LookUUID=EA1DE074F188405965EF399C72C221D9",
		"-XMP-crs:LookParametersClarity2012=+10",
		"-XMP-crs:LookParametersToneCurvePV2012=0, 0; 32, 22; 64, 56; 128, 128; 224, 232; 240, 246; 255, 255",
		"-XMP-crs:LookParametersLookTable=2FE663AB0D3CE5DA7B9F657BBCD66DFE",
	},
}

type makeModel struct{ make, model string }

var cameraProfilesMtx sync.Mutex
var cameraProfiles = map[makeModel]struct {
	adobe string
	other []string
}{}

func loadProfiles(make, model string, process float32, grayscale bool, profile, look string) (string, []string) {
	adobe, other := func() (string, []string) {
		cameraProfilesMtx.Lock()
		defer cameraProfilesMtx.Unlock()

		res, ok := cameraProfiles[makeModel{make, model}]
		if ok {
			return res.adobe, res.other
		}

		craw.EmbedProfiles = config.DngConverter
		profiles, _ := craw.GetCameraProfileNames(make, model)

		res.adobe = "Adobe Standard"
		for _, name := range profiles {
			if util.Contains(profiles, name+" v2") {
				// skip legacy
				continue
			}
			if strings.HasPrefix(name, "Adobe Standard") {
				res.adobe = name
			} else {
				res.other = append(res.other, string(name))
			}
		}

		sort.Strings(res.other)

		cameraProfiles[makeModel{make, model}] = res
		return res.adobe, res.other
	}()

	if process != 0 || profile != "" || look != "" {
		switch {
		case look == "" && (profile == "" || profile == adobe):
			profile = "Adobe Standard"
		case util.Contains(defaultProfiles, look) && (profile == "" || profile == adobe):
			profile = look
		case util.Contains(other, profile) && look == "" && !grayscale:
			//
		default:
			profile = "Custom"
		}
		if profile == "Adobe Standard" && grayscale {
			profile = "Adobe Standard B&W"
		}
	}

	return profile, other
}
