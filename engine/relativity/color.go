package relativity

import "math"

// RGB represents a color with 8-bit components.
type RGB struct {
	R, G, B uint8
}

// ShiftColorTemperature adjusts a color temperature by a Doppler factor.
// baseTemp: original temperature in Kelvin (e.g., 5800K for Sun)
// dopplerFactor: D > 1 blueshifts (higher temp), D < 1 redshifts (lower temp)
func ShiftColorTemperature(baseTemp, dopplerFactor float64) float64 {
	shifted := baseTemp * dopplerFactor
	// Clamp to reasonable star temperature range
	if shifted < 1000 {
		return 1000 // Very red, almost invisible
	}
	if shifted > 40000 {
		return 40000 // Very blue
	}
	return shifted
}

// TemperatureToRGB converts a blackbody temperature (Kelvin) to RGB.
// Uses Tanner Helland's approximation of the Planckian locus.
// Valid for temperatures 1000K - 40000K.
func TemperatureToRGB(temp float64) RGB {
	temp = clamp(temp, 1000, 40000)
	temp = temp / 100.0 // Algorithm uses temp/100

	var r, g, b float64

	// Red
	if temp <= 66 {
		r = 255
	} else {
		r = 329.698727446 * math.Pow(temp-60, -0.1332047592)
		r = clamp(r, 0, 255)
	}

	// Green
	if temp <= 66 {
		g = 99.4708025861*math.Log(temp) - 161.1195681661
	} else {
		g = 288.1221695283 * math.Pow(temp-60, -0.0755148492)
	}
	g = clamp(g, 0, 255)

	// Blue
	if temp >= 66 {
		b = 255
	} else if temp <= 19 {
		b = 0
	} else {
		b = 138.5177312231*math.Log(temp-10) - 305.0447927307
		b = clamp(b, 0, 255)
	}

	return RGB{uint8(r), uint8(g), uint8(b)}
}

// DopplerShiftColor applies Doppler shift to a base color.
// Estimates the original temperature from the color, shifts it, then converts back.
func DopplerShiftColor(baseColor RGB, dopplerFactor float64) RGB {
	// Estimate temperature from color (rough approximation)
	baseTemp := EstimateTemperature(baseColor)

	// Shift temperature
	shiftedTemp := ShiftColorTemperature(baseTemp, dopplerFactor)

	// Convert back to RGB
	return TemperatureToRGB(shiftedTemp)
}

// EstimateTemperature roughly estimates blackbody temperature from RGB.
// This is an approximation - colors don't map uniquely to temperatures.
func EstimateTemperature(c RGB) float64 {
	r := float64(c.R)
	b := float64(c.B)

	// Use blue/red ratio as rough temperature indicator
	if r < 1 {
		r = 1
	}
	ratio := b / r

	// Map ratio to temperature (very rough)
	// ratio ~0.3 -> ~3000K (red)
	// ratio ~1.0 -> ~6500K (white)
	// ratio ~1.5 -> ~10000K (blue-white)
	// ratio ~2.0 -> ~20000K (blue)
	if ratio < 0.5 {
		return 3000 + ratio*4000
	} else if ratio < 1.0 {
		return 5000 + (ratio-0.5)*3000
	} else if ratio < 1.5 {
		return 6500 + (ratio-1.0)*7000
	} else {
		return 10000 + (ratio-1.5)*20000
	}
}

// BeamBrightness computes relativistic beaming brightness factor.
// Intensity scales as D^3 where D is the Doppler factor.
func BeamBrightness(dopplerFactor float64) float64 {
	d := clamp(dopplerFactor, 0.01, 10.0) // Clamp to avoid extreme values
	return d * d * d
}

// ApplyBeaming adjusts an RGB color's brightness by beaming factor.
func ApplyBeaming(c RGB, beamFactor float64) RGB {
	// Clamp factor to reasonable range
	f := clamp(beamFactor, 0.0, 10.0)

	r := clamp(float64(c.R)*f, 0, 255)
	g := clamp(float64(c.G)*f, 0, 255)
	b := clamp(float64(c.B)*f, 0, 255)

	return RGB{uint8(r), uint8(g), uint8(b)}
}

// StarTemperature maps spectral class to approximate temperature.
var StarTemperature = map[string]float64{
	"O": 30000, // Blue
	"B": 20000, // Blue-white
	"A": 9000,  // White
	"F": 7000,  // Yellow-white
	"G": 5800,  // Yellow (Sun)
	"K": 4500,  // Orange
	"M": 3000,  // Red
}

// ProcessStar computes the apparent color and brightness of a star
// as seen from a relativistically moving observer.
func ProcessStar(baseTemp float64, direction Vec3, beta Vec3, gamma float64) (RGB, float64) {
	// Compute Doppler factor for this star's direction
	d := DopplerFactor(beta, direction, gamma)

	// Shift temperature
	shiftedTemp := ShiftColorTemperature(baseTemp, d)

	// Convert to color
	color := TemperatureToRGB(shiftedTemp)

	// Compute brightness factor
	brightness := BeamBrightness(d)

	return color, brightness
}
