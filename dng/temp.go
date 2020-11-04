package dng

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

func GetTemperature(x, y float64) (temperature, tint int) {
	tmp, tnt := xy64{x, y}.temperature()

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
