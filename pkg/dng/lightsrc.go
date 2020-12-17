// Copyright (c) 2020 Nuno Cruces
// SPDX-License-Identifier: MIT

package dng

// LightSource represents a kind of light source.
type LightSource uint16

// Values for the LightSource EXIF tag.
const (
	LSUnknown LightSource = iota
	LSDaylight
	LSFluorescent
	LSTungsten
	LSFlash
	_
	_
	_
	_
	LSFineWeather
	LSCloudyWeather
	LSShade
	LSDaylightFluorescent  // D  5700 - 7100K
	LSDayWhiteFluorescent  // N  4600 - 5500K
	LSCoolWhiteFluorescent // W  3800 - 4500K
	LSWhiteFluorescent     // WW 3250 - 3800K
	LSWarmWhiteFluorescent // L  2600 - 3250K
	LSStandardLightA
	LSStandardLightB
	LSStandardLightC
	LSD55
	LSD65
	LSD75
	LSD50
	LSISOStudioTungsten

	LSOther LightSource = 255
)

// Port of dng_camera_profile::IlluminantToTemperature

// Temperature gets the color temperature of the illuminant.
func (ls LightSource) Temperature() float64 {
	switch ls {

	case LSStandardLightA, LSTungsten:
		return 2850.0

	case LSISOStudioTungsten:
		return 3200.0

	case LSD50:
		return 5000.0

	case LSD55, LSDaylight, LSFineWeather, LSFlash, LSStandardLightB:
		return 5500.0

	case LSD65, LSStandardLightC, LSCloudyWeather:
		return 6500.0

	case LSD75, LSShade:
		return 7500.0

	case LSDaylightFluorescent:
		return (5700.0 + 7100.0) * 0.5

	case LSDayWhiteFluorescent:
		return (4600.0 + 5500.0) * 0.5

	case LSCoolWhiteFluorescent, LSFluorescent:
		return (3800.0 + 4500.0) * 0.5

	case LSWhiteFluorescent:
		return (3250.0 + 3800.0) * 0.5

	case LSWarmWhiteFluorescent:
		return (2600.0 + 3250.0) * 0.5

	default:
		return 0.0
	}
}
