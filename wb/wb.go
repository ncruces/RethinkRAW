package wb

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

func (p *CameraProfile) GetTemperature(neutral []float64) (temperature, tint int) {
	vec := mat.NewVecDense(len(neutral), neutral)
	tmp, tnt := p.neutralToXY(vec).temperature()

	// temperature range 2000 to 50000
	switch {
	case tmp < 2000.0:
		tmp = 2000.0
	case tmp > 50000.0:
		tmp = 50000.0
	default:
		tmp = math.RoundToEven(tmp)
	}

	// tint range -150 to +150
	switch {
	case tnt < -150.0:
		tnt = -150.0
	case tint > +150.0:
		tnt = +150.0
	default:
		tnt = math.RoundToEven(tnt)
	}

	return int(tmp), int(tnt)
}

// Port of dng_color_spec::dng_color_spec.
func (p *CameraProfile) setup() {
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
		p.setup()
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

var _D50 = xy64{0.34567, 0.35850}

type xy64 struct{ x, y float64 }
type xyz64 struct{ x, y, z float64 }

func newXYZ64(v mat.Vector) xyz64 {
	if v.Len() != 3 {
		panic(mat.ErrShape)
	}
	return xyz64{v.AtVec(0), v.AtVec(1), v.AtVec(2)}
}

// Port of XYZtoXY.
func (v xyz64) XY() xy64 {
	total := v.x + v.y + v.z
	if total <= 0.0 {
		return _D50
	}
	return xy64{v.x / total, v.y / total}
}

// Scale factor between distances in uv space to a more user friendly "tint"
// parameter.
const tintScale = -3000.0

// Table from Wyszecki & Stiles, "Color Science", second edition, page 228.
var tempTable = [31]struct {
	r, u, v, t float64
}{
	{0, 0.18006, 0.26352, -0.24341},
	{10, 0.18066, 0.26589, -0.25479},
	{20, 0.18133, 0.26846, -0.26876},
	{30, 0.18208, 0.27119, -0.28539},
	{40, 0.18293, 0.27407, -0.30470},
	{50, 0.18388, 0.27709, -0.32675},
	{60, 0.18494, 0.28021, -0.35156},
	{70, 0.18611, 0.28342, -0.37915},
	{80, 0.18740, 0.28668, -0.40955},
	{90, 0.18880, 0.28997, -0.44278},
	{100, 0.19032, 0.29326, -0.47888},
	{125, 0.19462, 0.30141, -0.58204},
	{150, 0.19962, 0.30921, -0.70471},
	{175, 0.20525, 0.31647, -0.84901},
	{200, 0.21142, 0.32312, -1.0182},
	{225, 0.21807, 0.32909, -1.2168},
	{250, 0.22511, 0.33439, -1.4512},
	{275, 0.23247, 0.33904, -1.7298},
	{300, 0.24010, 0.34308, -2.0637},
	{325, 0.24792, 0.34655, -2.4681}, /* Note: 0.24792 is a corrected value for the error found in W&S as 0.24702 */
	{350, 0.25591, 0.34951, -2.9641},
	{375, 0.26400, 0.35200, -3.5814},
	{400, 0.27218, 0.35407, -4.3633},
	{425, 0.28039, 0.35577, -5.3762},
	{450, 0.28863, 0.35714, -6.7262},
	{475, 0.29685, 0.35823, -8.5955},
	{500, 0.30505, 0.35907, -11.324},
	{525, 0.31320, 0.35968, -15.628},
	{550, 0.32129, 0.36011, -23.325},
	{575, 0.32931, 0.36038, -40.770},
	{600, 0.33724, 0.36051, -116.45},
}

// Port dng_temperature::Set_xy_coord.
func (xy xy64) temperature() (temperature, tint float64) {
	// Convert to uv space.
	u := 2.0 * xy.x / (1.5 - xy.x + 6.0*xy.y)
	v := 3.0 * xy.y / (1.5 - xy.x + 6.0*xy.y)

	// Search for line pair coordinate is between.
	last_dt := 0.0
	last_du := 0.0
	last_dv := 0.0

	for index := 1; index < len(tempTable); index++ {
		// Convert slope to delta-u and delta-v, with length 1.
		du := 1.0
		dv := tempTable[index].t
		{
			len := math.Hypot(du, dv)
			du /= len
			dv /= len
		}

		// Find delta from black body point to test coordinate.
		uu := u - tempTable[index].u
		vv := v - tempTable[index].v

		// Find distance above or below line.
		dt := -uu*dv + vv*du

		// If below line, we have found line pair.
		if dt <= 0.0 || index == len(tempTable)-1 {
			// Find fractional weight of two lines.
			if dt > 0.0 {
				dt = 0.0
			}
			dt = -dt

			f := 0.0
			if index != 1 {
				f = dt / (last_dt + dt)
			}

			// Interpolate the temperature.
			temperature = 1.0e6 / (tempTable[index-1].r*f + tempTable[index].r*(1.0-f))

			// Find delta from black body point to test coordinate.
			uu = u - (tempTable[index-1].u*f + tempTable[index].u*(1.0-f))
			vv = v - (tempTable[index-1].v*f + tempTable[index].v*(1.0-f))

			// Interpolate vectors along slope.
			du = du*(1.0-f) + last_du*f
			dv = dv*(1.0-f) + last_dv*f
			{
				len := math.Hypot(du, dv)
				du /= len
				dv /= len
			}

			// Find distance along slope.
			tint = (uu*du + vv*dv) * tintScale
			break
		}

		// Try next line pair.
		last_dt = dt
		last_du = du
		last_dv = dv
	}

	return temperature, tint
}

type LightSource uint16

// Values for LightSource tag.
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
