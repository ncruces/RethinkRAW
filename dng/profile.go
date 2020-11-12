package dng

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

type CameraProfile struct {
	CalibrationIlluminant1, CalibrationIlluminant2 LightSource
	ColorMatrix1, ColorMatrix2                     []float64
	CameraCalibration1, CameraCalibration2         []float64
	AnalogBalance                                  []float64

	temperature1, temperature2 float64
	colorMatrix1, colorMatrix2 *mat.Dense
}

func (p *CameraProfile) Init() error {
	return mat.Maybe(p.init)
}

func (p *CameraProfile) GetTemperature(neutral []float64) (temperature, tint int, err error) {
	err = mat.Maybe(func() {
		xy := p.neutralToXY(mat.NewVecDense(len(neutral), neutral))
		temperature, tint = GetTemperatureFromXY(xy.x, xy.y)
	})
	return
}

// Port of dng_color_spec::dng_color_spec.
func (p *CameraProfile) init() {
	channels := len(p.ColorMatrix1) / 3

	p.temperature1 = p.CalibrationIlluminant1.Temperature()
	p.temperature2 = p.CalibrationIlluminant2.Temperature()

	var analog *mat.DiagDense
	if p.AnalogBalance != nil {
		analog = mat.NewDiagDense(channels, p.AnalogBalance)
	}

	p.colorMatrix1 = mat.NewDense(channels, 3, p.ColorMatrix1)
	if p.CameraCalibration1 != nil {
		p.colorMatrix1.Mul(mat.NewDense(channels, channels, p.CameraCalibration1), p.colorMatrix1)
	}
	if analog != nil {
		p.colorMatrix1.Mul(analog, p.colorMatrix1)
	}

	if p.CameraCalibration2 == nil ||
		p.temperature1 == p.temperature2 ||
		p.temperature1 <= 0.0 || p.temperature2 <= 0.0 {
		p.temperature1 = 5000.0
		p.temperature2 = 5000.0
		p.colorMatrix2 = p.colorMatrix1
	} else {
		p.colorMatrix2 = mat.NewDense(channels, 3, p.ColorMatrix2)
		if p.CameraCalibration2 != nil {
			p.colorMatrix2.Mul(mat.NewDense(channels, channels, p.CameraCalibration2), p.colorMatrix2)
		}
		if analog != nil {
			p.colorMatrix2.Mul(analog, p.colorMatrix2)
		}

		if p.temperature1 > p.temperature2 {
			p.temperature1, p.temperature2 = p.temperature2, p.temperature1
			p.colorMatrix1, p.colorMatrix2 = p.colorMatrix2, p.colorMatrix1
		}
	}
}

// Port of dng_color_spec::NeutralToXY.
func (p *CameraProfile) neutralToXY(neutral mat.Vector) xy64 {
	const maxPasses = 30

	if neutral.Len() == 1 {
		return _D50
	}

	last := _D50
	for pass := 0; pass < maxPasses; pass++ {
		xyzToCamera := p.findXYZtoCamera(last)

		var vec mat.VecDense
		vec.SolveVec(xyzToCamera, neutral)
		next := newXYZ64(&vec).XY()

		if math.Abs(next.x-last.x)+math.Abs(next.y-last.y) < 0.0000001 {
			return next
		}

		// If we reach the limit without converging, we are most likely
		// in a two value oscillation.  So take the average of the last
		// two estimates and give up.
		if pass == maxPasses-1 {
			next.x = (last.x + next.x) * 0.5
			next.y = (last.y + next.y) * 0.5
		}
		last = next
	}
	return last
}

// Port of dng_color_spec::FindXYZtoCamera.
func (p *CameraProfile) findXYZtoCamera(white xy64) mat.Matrix {
	if p.colorMatrix1 == nil {
		p.init()
	}

	// Convert to temperature/offset space.
	temperature, _ := white.temperature()

	// Find fraction to weight the first calibration.
	var g float64
	if temperature <= p.temperature1 {
		g = 1.0
	} else if temperature >= p.temperature2 {
		g = 0.0
	} else {
		g = (1.0/temperature - 1.0/p.temperature2) /
			(1.0/p.temperature1 - 1.0/p.temperature2)
	}

	// Interpolate the color matrix.
	var colorMatrix mat.Dense

	if g >= 1.0 {
		return p.colorMatrix1
	} else if g <= 0.0 {
		return p.colorMatrix2
	} else {
		var c1, c2 mat.Dense
		c1.Scale(g, p.colorMatrix1)
		c2.Scale((1.0 - g), p.colorMatrix2)
		colorMatrix.Add(&c1, &c2)
	}

	// Return the interpolated color matrix.
	return &colorMatrix
}
