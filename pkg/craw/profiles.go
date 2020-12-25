package craw

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/rethinkraw/pkg/dng"
)

// GetCameraProfiles gets all the profiles that apply to a given camera.
// It looks for profiles under the GlobalSettings and UserSettings directories.
// Returns a list of paths to DCP files for the profiles.
func GetCameraProfiles(make, model string) ([]string, error) {
	glb, err := LoadIndex(filepath.Join(GlobalSettings, filepath.FromSlash("CameraProfiles/Index.dat")))
	if err != nil {
		return nil, err
	}
	usr, err := LoadIndex(filepath.Join(UserSettings, filepath.FromSlash("CameraProfiles/Index.dat")))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	make = strings.ToUpper(make)
	model = strings.ToUpper(model)
	makeModel := make + " " + model

	var profiles []string
	for _, rec := range append(glb, usr...) {
		camera := strings.ToUpper(rec.Prop["model_restriction"])
		if camera == model || camera == makeModel {
			profile, err := dng.GetDCPProfileName(rec.Path)
			if err != nil {
				return nil, err
			}
			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}
