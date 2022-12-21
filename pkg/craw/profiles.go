package craw

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/ncruces/rethinkraw/pkg/dng"
)

// GetCameraProfiles gets all the profiles that apply to a given camera.
// Returns the DCP file paths for the profiles.
// It looks for profiles under the GlobalSettings and UserSettings directories.
func GetCameraProfiles(make, model string) ([]string, error) {
	once.Do(initPaths)

	glb, err := LoadIndex(filepath.Join(GlobalSettings, filepath.FromSlash("CameraProfiles/Index.dat")))
	if err != nil {
		return nil, err
	}
	usr, err := LoadIndex(filepath.Join(UserSettings, filepath.FromSlash("CameraProfiles/Index.dat")))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	make = strings.ToUpper(make)
	model = strings.ToUpper(model)
	makeModel := make + " " + model

	var profiles []string
	for _, rec := range append(glb, usr...) {
		test := strings.ToUpper(rec.Prop["model_restriction"])
		var matches bool
		if test == "" || test == model || test == makeModel {
			matches = true
		} else if testMake, testModel, ok := strings.Cut(test, " "); ok {
			matches = testModel == model && strings.Contains(make, testMake)
		}
		if matches {
			profiles = append(profiles, rec.Path)
		}
	}

	return profiles, nil
}

// GetCameraProfileNames gets the names of all profiles that apply to a given camera.
// It looks for profiles under the GlobalSettings and UserSettings directories,
// and in the EmbedProfiles file, if set.
func GetCameraProfileNames(make, model string) ([]string, error) {
	once.Do(initPaths)

	profiles, err := GetCameraProfiles(make, model)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, profile := range profiles {
		name, err := dng.GetDCPProfileName(profile)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	if make == "FUJIFILM" {
		embed, err := fujiCameraProfiles(model)
		if err != nil {
			return nil, err
		}
		names = append(names, embed...)
	}

	return names, nil
}
