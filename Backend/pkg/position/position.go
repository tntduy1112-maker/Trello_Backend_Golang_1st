package position

import "math"

const (
	InitialGap   = 65536.0
	MinGap       = 1.0
	RebalanceGap = 65536.0
)

func NextPosition(currentMax float64) float64 {
	if currentMax == 0 {
		return InitialGap
	}
	return currentMax + InitialGap
}

func Between(before, after float64) float64 {
	return (before + after) / 2
}

func BeforeFirst(firstPosition float64) float64 {
	return firstPosition / 2
}

func NeedsRebalance(before, after float64) bool {
	return math.Abs(after-before) < MinGap
}

func Rebalance(count int) []float64 {
	positions := make([]float64, count)
	for i := 0; i < count; i++ {
		positions[i] = float64(i+1) * RebalanceGap
	}
	return positions
}
