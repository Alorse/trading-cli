package utils

// RoundTo rounds a float64 to n decimal places.
func RoundTo(v float64, n int) float64 {
	if n <= 0 {
		return v
	}
	pow := 1.0
	for i := 0; i < n; i++ {
		pow *= 10
	}
	return float64(int64(v*pow+0.5)) / pow
}

// Between returns true if v is in [min, max].
func Between(v, min, max float64) bool {
	return v >= min && v <= max
}

// Clamp restricts v to [min, max].
func Clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// SafeDivide divides a by b, returning 0 if b is 0.
func SafeDivide(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

// Max returns the larger of two float64 values.
func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Min returns the smaller of two float64 values.
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Abs returns the absolute value.
func Abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// PercentChange computes the percentage change from old to new.
func PercentChange(old, new float64) float64 {
	return SafeDivide(new-old, old) * 100
}
