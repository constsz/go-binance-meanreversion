package ta

import "fmt"

func IndicatorName(indicatorType string, strategyId int) string {
	switch indicatorType {
	case "BetterBands":
		return fmt.Sprintf("bb-%d", strategyId)
	}
	return ""
}

// Pct converts human readable percentage to multiplier.
func Pct(humanReadablePercentage float64) float64 {
	return humanReadablePercentage/100 + 1
}

func PercentsOfNumber(percentage float64, number float64) float64 {
	return (number / 100) * percentage
}

func PctChangeMult(a, b float64) float64 {
	return (PctChange(a, b) / 100) + 1
}

// PctChange Returns human-readable percent difference (100 and 97 returns -3%)
func PctChange(a, b float64) float64 {
	return ((b - a) / a) * 100
}

func PercentsXofY(min, max float64) float64 {
	return min / (max / 100)
}
