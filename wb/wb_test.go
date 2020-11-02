package wb_test

import (
	"rethinkraw/wb"
	"testing"
)

func TestCameraProfile_GetTemperature(t *testing.T) {
	// For a RG/GB camera.
	{
		cam := wb.CameraProfile{
			CalibrationIlluminant1: wb.LSStandardLightA,
			CalibrationIlluminant2: wb.LSD65,
			ColorMatrix1:           []float64{0.9210, -0.4777, +0.0345, -0.4492, 1.3117, 0.1471, -0.0345, 0.0879, 0.6708},
			ColorMatrix2:           []float64{0.7657, -0.2847, -0.0607, -0.4083, 1.1966, 0.2389, -0.0684, 0.1418, 0.5844},
			CameraCalibration1:     []float64{0.9434, 0, 0, 0, 1, 0, 0, 0, 0.94},
			CameraCalibration2:     []float64{0.9434, 0, 0, 0, 1, 0, 0, 0, 0.94},
		}

		temp, tint := cam.GetTemperature([]float64{0.346414, 1, 0.636816})
		if temp != 6383 || tint != 1 {
			t.Error(temp, tint)
		}
	}

	// For a 4 color RGB+E camera (F828).
	// Multipliers calculated with dcraw.
	{
		cam := wb.CameraProfile{
			CalibrationIlluminant1: wb.LSStandardLightA,
			CalibrationIlluminant2: wb.LSD65,
			ColorMatrix1:           []float64{0.8771, -0.3148, -0.0125, -0.5926, 1.2567, 0.3815, -0.0871, 0.1575, 0.6633, -0.4678, 0.8486, 0.4548},
			ColorMatrix2:           []float64{0.7925, -0.1910, -0.0776, -0.8227, 1.5459, 0.2998, -0.1517, 0.2198, 0.6817, -0.7241, 1.1401, 0.3481},
			AnalogBalance:          []float64{1, 1, 1, 1},
		}

		temp, tint := cam.GetTemperature([]float64{1 / 1.080806, 1, 1 / 3.700866, 1 / 1.623588})
		if temp != 2681 || tint != 28 {
			t.Error(temp, tint)
		}
	}
}
