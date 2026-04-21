package indicators

import (
	"math"
)

// floatEqual compares two floats with tolerance for floating point precision.
func floatEqual(a, b float64) bool {
	tolerance := 1e-3
	return math.Abs(a-b) < tolerance
}
