package main

import (
	"bytes"
	"log"
	"strings"
	"sync"

	"rethinkraw/craw"
)

var defaultProfiles = []string{
	"Adobe Color", "Adobe Monochrome", "Adobe Landscape", "Adobe Neutral",
	"Adobe Portrait", "Adobe Vivid", "Adobe Standard", "Adobe Standard B&W",
}

var profileSettings = map[string][]string{
	"Adobe Standard": {
		"-XMP-crs:ConvertToGrayscale=False",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
	},
	"Adobe Standard B&W": {
		"-XMP-crs:ConvertToGrayscale=True",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
	},
	"Adobe Color": {
		"-XMP-crs:ConvertToGrayscale=False",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Color",
		"-XMP-crs:LookUUID=B952C231111CD8E0ECCF14B86BAA7077",
		"-XMP-crs:LookParametersToneCurvePV2012+=0, 0",
		"-XMP-crs:LookParametersToneCurvePV2012+=22, 16",
		"-XMP-crs:LookParametersToneCurvePV2012+=40, 35",
		"-XMP-crs:LookParametersToneCurvePV2012+=127, 127",
		"-XMP-crs:LookParametersToneCurvePV2012+=224, 230",
		"-XMP-crs:LookParametersToneCurvePV2012+=240, 246",
		"-XMP-crs:LookParametersToneCurvePV2012+=255, 255",
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
		"-XMP-crs:LookParametersToneCurvePV2012+=0, 0",
		"-XMP-crs:LookParametersToneCurvePV2012+=64, 56",
		"-XMP-crs:LookParametersToneCurvePV2012+=128, 128",
		"-XMP-crs:LookParametersToneCurvePV2012+=192, 197",
		"-XMP-crs:LookParametersToneCurvePV2012+=255, 255",
		"-XMP-crs:LookParametersLookTable=73ED6C18DDE909DD7EA2D771F5AC282D",
	},
	"Adobe Landscape": {
		"-XMP-crs:ConvertToGrayscale=False",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Landscape",
		"-XMP-crs:LookUUID=6F9C877E84273F4E8271E6B91BEB36A1",
		"-XMP-crs:LookParametersHighlights2012=-12",
		"-XMP-crs:LookParametersShadows2012=+12",
		"-XMP-crs:LookParametersClarity2012=+10",
		"-XMP-crs:LookParametersToneCurvePV2012+=0, 0",
		"-XMP-crs:LookParametersToneCurvePV2012+=64, 60",
		"-XMP-crs:LookParametersToneCurvePV2012+=128, 128",
		"-XMP-crs:LookParametersToneCurvePV2012+=192, 196",
		"-XMP-crs:LookParametersToneCurvePV2012+=255, 255",
		"-XMP-crs:LookParametersLookTable=0B3BFB5CFB7DBF7FF175E98F24D316B0",
	},
	"Adobe Neutral": {
		"-XMP-crs:ConvertToGrayscale=False",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Neutral",
		"-XMP-crs:LookUUID=1E8E067A11CD44394A3C36A327BB34D1",
		"-XMP-crs:LookParametersToneCurvePV2012+=0, 0",
		"-XMP-crs:LookParametersToneCurvePV2012+=16, 24",
		"-XMP-crs:LookParametersToneCurvePV2012+=64, 72",
		"-XMP-crs:LookParametersToneCurvePV2012+=128, 128",
		"-XMP-crs:LookParametersToneCurvePV2012+=192, 176",
		"-XMP-crs:LookParametersToneCurvePV2012+=244, 234",
		"-XMP-crs:LookParametersToneCurvePV2012+=255, 255",
		"-XMP-crs:LookParametersLookTable=7740BB918B2F6D93D7B95A4DBB78DB23",
	},
	"Adobe Portrait": {
		"-XMP-crs:ConvertToGrayscale=False",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Portrait",
		"-XMP-crs:LookUUID=D6496412E06A83789C499DF9540AA616",
		"-XMP-crs:LookParametersToneCurvePV2012+=0, 0",
		"-XMP-crs:LookParametersToneCurvePV2012+=66, 64",
		"-XMP-crs:LookParametersToneCurvePV2012+=190, 192",
		"-XMP-crs:LookParametersToneCurvePV2012+=255, 255",
		"-XMP-crs:LookParametersLookTable=E5A76DBB8B3F132A04C01AF45DC2EF1B",
	},
	"Adobe Vivid": {
		"-XMP-crs:ConvertToGrayscale=False",
		"-XMP-crs:CameraProfile=",
		"-XMP-crs:Look*=",
		"-XMP-crs:LookName=Adobe Vivid",
		"-XMP-crs:LookUUID=EA1DE074F188405965EF399C72C221D9",
		"-XMP-crs:LookParametersClarity2012=+10",
		"-XMP-crs:LookParametersToneCurvePV2012+=0, 0",
		"-XMP-crs:LookParametersToneCurvePV2012+=32, 22",
		"-XMP-crs:LookParametersToneCurvePV2012+=64, 56",
		"-XMP-crs:LookParametersToneCurvePV2012+=128, 128",
		"-XMP-crs:LookParametersToneCurvePV2012+=224, 232",
		"-XMP-crs:LookParametersToneCurvePV2012+=240, 246",
		"-XMP-crs:LookParametersToneCurvePV2012+=255, 255",
		"-XMP-crs:LookParametersLookTable=2FE663AB0D3CE5DA7B9F657BBCD66DFE",
	},
}

var cameraProfilesMtx sync.Mutex
var cameraProfiles = map[string]struct {
	adobe string
	other []string
}{}

func getCameraProfiles(make, model string) (string, []string) {
	cameraProfilesMtx.Lock()
	defer cameraProfilesMtx.Unlock()

	res, ok := cameraProfiles[make+" "+model]
	if ok {
		return res.adobe, res.other
	}

	paths, _ := craw.GetCameraProfiles(make, model)
	for _, path := range paths {
		log.Print("exiftool (load profile)...")
		n, _ := exifserver.Command("--printConv", "-short3", "-fast2", "-ProfileName", path)
		name := string(bytes.TrimSpace(n))
		if strings.HasPrefix(name, "Adobe Standard") {
			res.adobe = name
		} else {
			res.other = append(res.other, string(name))
		}
	}
	cameraProfiles[make+" "+model] = res
	return res.adobe, res.other
}
