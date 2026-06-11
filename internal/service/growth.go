package service

// CalculateGrowth computes the new ability value after applying a boost.
// boostPercentage comes from the activity effect (e.g., 1.0 = +1%).
// growthRate comes from the ability (default 1.0, multiplier on the boost).
func CalculateGrowth(currentValue, boostPercentage, growthRate float64) float64 {
	effectiveBoost := boostPercentage * growthRate
	return currentValue * (1 + effectiveBoost/100)
}
