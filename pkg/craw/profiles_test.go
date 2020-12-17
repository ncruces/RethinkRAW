// Copyright (c) 2020 Nuno Cruces
// SPDX-License-Identifier: MIT

package craw

import (
	"testing"
)

func TestGetCameraProfiles(t *testing.T) {
	profiles, err := GetCameraProfiles("SONY", "ILCE-7")
	if err != nil {
		t.Error(err)
	} else if len(profiles) != 9 {
		t.Errorf("Expected 9 profiles got %d", len(profiles))
	}
}
