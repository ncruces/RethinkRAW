package craw

import (
	"os"
	"path/filepath"
	"strings"
)

// GetCameraProfiles gets all the profiles that apply to a given camera.
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
			profiles = append(profiles, rec.Path)
		}
	}

	return profiles, nil
}
