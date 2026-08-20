package stable

import "math"

// PopulationStdDev computes population standard deviation for values.
func PopulationStdDev(values []int64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += float64(v)
	}
	mean := sum / float64(n)
	var sumSq float64
	for _, v := range values {
		diff := float64(v) - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(n))
}

// MeanInt64 returns arithmetic mean of int64 slice.
func MeanInt64(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += float64(v)
	}
	return sum / float64(len(values))
}

// MinMax returns minimum and maximum values in slice.
func MinMax(values []int64) (min, max int64) {
	if len(values) == 0 {
		return 0, 0
	}
	min = values[0]
	max = values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// Range returns max - min for values.
func Range(values []int64) int64 {
	min, max := MinMax(values)
	return max - min
}

// IsStable checks whether values meet full-window stability criteria.
func IsStable(values []int64, expectedN int, eps float64) bool {
	if len(values) != expectedN {
		return false
	}
	return PopulationStdDev(values) <= eps
}
